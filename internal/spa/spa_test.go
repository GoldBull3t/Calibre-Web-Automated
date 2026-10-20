package spa

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSPAHandler(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatalf("Erro ao inicializar SPA: %v", err)
	}

	if !h.HasAssets() {
		t.Skip("Assets do frontend ainda não compilados")
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado código 200, recebido: %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<!doctype html>") && !strings.Contains(body, "<html") {
		t.Errorf("Conteúdo não parece HTML válido: %s", body)
	}
}

