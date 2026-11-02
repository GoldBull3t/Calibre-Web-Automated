package formats

import (
	"context"
	"log"
	"sync"
)

// TaskFn representa uma tarefa assíncrona de conversão ou otimização.
type TaskFn func(ctx context.Context) error

// ConverterPool é uma fila de trabalhadores em Go com limitação de concorrência (max_workers).
type ConverterPool struct {
	maxWorkers int
	taskQueue  chan TaskFn
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewConverterPool cria uma nova pool de trabalhadores assíncronos.
func NewConverterPool(maxWorkers, queueSize int) *ConverterPool {
	if maxWorkers <= 0 {
		maxWorkers = 2
	}
	if queueSize <= 0 {
		queueSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())
	pool := &ConverterPool{
		maxWorkers: maxWorkers,
		taskQueue:  make(chan TaskFn, queueSize),
		ctx:        ctx,
		cancel:     cancel,
	}

	pool.start()
	return pool
}

func (p *ConverterPool) start() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()
			for {
				select {
				case <-p.ctx.Done():
					return
				case task, ok := <-p.taskQueue:
					if !ok {
						return
					}
					if err := task(p.ctx); err != nil {
						log.Printf("[WORKER %d ERROR] Falha na tarefa de conversão: %v", workerID, err)
					}
				}
			}
		}(i)
	}
}

// Submit submete uma tarefa para processamento assíncrono não-bloqueante.
func (p *ConverterPool) Submit(task TaskFn) bool {
	if p == nil {
		return false
	}
	select {
	case p.taskQueue <- task:
		return true
	default:
		// Fila cheia, descarta ou executa conforme política
		log.Printf("[CONVERTER POOL] Fila cheia, a descartar tarefa")
		return false
	}
}

// Close aguarda a conclusão das tarefas ativas e encerra os trabalhadores.
func (p *ConverterPool) Close() {
	if p == nil {
		return
	}
	p.cancel()
	close(p.taskQueue)
	p.wg.Wait()
}
