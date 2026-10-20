package events

import (
	"testing"
	"time"
)

func TestEventBroker(t *testing.T) {
	broker := GetBroker()

	ch, unsubscribe := broker.Subscribe()
	defer unsubscribe()

	testMsg := "Novo livro importado com sucesso"
	broker.Broadcast("book.ingested", testMsg, map[string]any{"id": 42})

	select {
	case ev := <-ch:
		if ev.Type != "book.ingested" || ev.Message != testMsg {
			t.Fatalf("Evento inesperado recebido: %+v", ev)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout à espera do evento broadcast")
	}
}

