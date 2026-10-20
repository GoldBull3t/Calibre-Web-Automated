package cache

import (
	"testing"
)

func TestByteLRUCache(t *testing.T) {
	// Cache com 100 bytes máximo
	c := NewByteLRUCache(100)

	c.Put("item1", make([]byte, 40), "image/jpeg")
	c.Put("item2", make([]byte, 40), "image/png")

	count, bytesUsed, max := c.Stats()
	if count != 2 || bytesUsed != 80 || max != 100 {
		t.Fatalf("Esperado 2 itens e 80 bytes, obtido %d itens e %d bytes", count, bytesUsed)
	}

	// Adiciona item3 com 40 bytes -> deve ejetar item1 (LRU)
	c.Put("item3", make([]byte, 40), "image/webp")

	if _, _, ok := c.Get("item1"); ok {
		t.Fatal("Esperado que item1 fosse ejetado por exceder capacidade")
	}

	if _, mime, ok := c.Get("item2"); !ok || mime != "image/png" {
		t.Fatal("Esperado que item2 continue na cache")
	}

	if _, mime, ok := c.Get("item3"); !ok || mime != "image/webp" {
		t.Fatal("Esperado que item3 continue na cache")
	}

	c.Delete("item2")
	if _, _, ok := c.Get("item2"); ok {
		t.Fatal("Esperado que item2 fosse apagado")
	}

	c.Clear()
	if count, bytesUsed, _ := c.Stats(); count != 0 || bytesUsed != 0 {
		t.Fatal("Esperado cache limpa após Clear()")
	}
}

