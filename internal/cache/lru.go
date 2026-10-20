package cache

import (
	"container/list"
	"sync"
)

type cacheEntry struct {
	key      string
	data     []byte
	mimeType string
	size     int
}

// ByteLRUCache é uma cache LRU thread-safe com limite de bytes em memória.
type ByteLRUCache struct {
	mu        sync.Mutex
	maxBytes  int
	curBytes  int
	items     map[string]*list.Element
	evictList *list.List
}

// NewByteLRUCache inicializa uma nova cache LRU com o tamanho máximo em bytes especificado.
func NewByteLRUCache(maxBytes int) *ByteLRUCache {
	return &ByteLRUCache{
		maxBytes:  maxBytes,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

// Get obtém os bytes e o mimeType de um item da cache, movendo-o para a frente da lista.
func (c *ByteLRUCache) Get(key string) ([]byte, string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		return entry.data, entry.mimeType, true
	}
	return nil, "", false
}

// Put insere ou atualiza um item na cache, removendo os itens mais antigos se exceder o limite de bytes.
func (c *ByteLRUCache) Put(key string, data []byte, mimeType string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	size := len(data)
	if size > c.maxBytes {
		return // Não armazena itens individuais que excedam o tamanho total da cache
	}

	// Se já existe, atualiza
	if elem, ok := c.items[key]; ok {
		entry := elem.Value.(*cacheEntry)
		c.curBytes -= entry.size
		entry.data = data
		entry.mimeType = mimeType
		entry.size = size
		c.curBytes += size
		c.evictList.MoveToFront(elem)
	} else {
		entry := &cacheEntry{
			key:      key,
			data:     data,
			mimeType: mimeType,
			size:     size,
		}
		elem := c.evictList.PushFront(entry)
		c.items[key] = elem
		c.curBytes += size
	}

	// Ejeção dos itens mais antigos até estar dentro do limite
	for c.curBytes > c.maxBytes && c.evictList.Len() > 0 {
		oldest := c.evictList.Back()
		if oldest != nil {
			c.removeElement(oldest)
		}
	}
}

// Delete remove um item da cache.
func (c *ByteLRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.removeElement(elem)
	}
}

// Clear limpa toda a cache.
func (c *ByteLRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.evictList.Init()
	c.curBytes = 0
}

func (c *ByteLRUCache) removeElement(elem *list.Element) {
	c.evictList.Remove(elem)
	entry := elem.Value.(*cacheEntry)
	delete(c.items, entry.key)
	c.curBytes -= entry.size
}

// Stats devolve estatísticas de ocupação da cache.
func (c *ByteLRUCache) Stats() (int, int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.evictList.Len(), c.curBytes, c.maxBytes
}

