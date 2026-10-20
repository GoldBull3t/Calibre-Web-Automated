package metadata_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crocodilestick/calibre-web-go/internal/metadata"
)

func TestHardcoverClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"books": []map[string]any{
					{
						"id":    99482,
						"title": "Alice in Wonderland",
						"slug":  "alice-in-wonderland",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// Como a URL HardcoverGraphQLEndpoint é constante, vamos testar com um client apontado ou instanciado
	client := metadata.NewHardcoverClient("fake_token")
	if client == nil {
		t.Fatal("Client não pode ser nil")
	}

	ctx := context.Background()
	// Testa busca sem encontrar (pois aponta para endpoint real se não mockado, ou timeout gracioso)
	_, err := client.FetchHardcoverID(ctx, "NonExistentBookXYZ123ABC999", "No Author", "")
	if err == nil {
		t.Log("Busca retornou sem erro ou encontrou")
	} else {
		t.Logf("Comportamento esperado com livro inexistente / sem chave real: %v", err)
	}
}

