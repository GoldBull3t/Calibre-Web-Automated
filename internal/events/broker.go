package events

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Event representa uma mensagem SSE enviada aos clientes conectados em tempo real.
type Event struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Payload   any    `json:"payload,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// Broker gere os canais de subscrição de clientes SSE.
type Broker struct {
	mu      sync.RWMutex
	clients map[chan Event]struct{}
}

var (
	defaultBroker *Broker
	once          sync.Once
)

// GetBroker devolve a instância singleton do Broker SSE.
func GetBroker() *Broker {
	once.Do(func() {
		defaultBroker = &Broker{
			clients: make(map[chan Event]struct{}),
		}
	})
	return defaultBroker
}

// Subscribe regista um novo canal de receção de eventos para um cliente.
func (b *Broker) Subscribe() (chan Event, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, 32)
	b.clients[ch] = struct{}{}

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if _, exists := b.clients[ch]; exists {
			delete(b.clients, ch)
			close(ch)
		}
	}

	return ch, unsubscribe
}

// Broadcast emite um evento para todos os clientes conectados.
func (b *Broker) Broadcast(eventType, message string, payload any) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.clients) == 0 {
		return
	}

	ev := Event{
		Type:      eventType,
		Message:   message,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}

	for ch := range b.clients {
		select {
		case ch <- ev:
		default:
			// Não bloqueia se o cliente for lento a consumir
		}
	}
}

// ServeHTTP entrega o stream de Server-Sent Events ao cliente HTTP.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming SSE não suportado pelo servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	eventChan, unsubscribe := b.Subscribe()
	defer unsubscribe()

	// Envia ping inicial de confirmação de conexão
	_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	ticker := time.NewTicker(25 * time.Second) // Keep-alive heartbeat
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			_, err := fmt.Fprintf(w, ": heartbeat\n\n")
			if err != nil {
				return
			}
			flusher.Flush()
		case ev, ok := <-eventChan:
			if !ok {
				return
			}
			data, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, string(data))
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

