package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	// 5 tokens max, 1 token/sec
	rl := NewRateLimiter(1, 3)

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 3 requisições consecutivas devem passar (burst = 3)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.50:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("Requisição %d deveria ter passado, status: %d", i+1, w.Code)
		}
	}

	// 4ª requisição imediata deve ser bloqueada com 429 Too Many Requests
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.50:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("Esperado 429 Too Many Requests, obtido: %d", w.Code)
	}

	// Outro IP deve passar normalmente
	reqOther := httptest.NewRequest("GET", "/test", nil)
	reqOther.RemoteAddr = "10.0.0.1:5678"
	wOther := httptest.NewRecorder()
	handler.ServeHTTP(wOther, reqOther)
	if wOther.Code != http.StatusOK {
		t.Fatalf("Requisição de IP diferente deveria ter passado, status: %d", wOther.Code)
	}
}

