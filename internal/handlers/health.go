package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/database"
)

// Healthz é a sonda de liveness simples (HTTP 200).
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready é a sonda de prontidão profunda que valida a integridade ativa do SQLite.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := h.db.HealthCheck(r.Context())

	if status.Status == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(status)
}

// APIQuarantine lista ou gere itens em quarentena.
func (h *Handler) APIQuarantine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	filter := r.URL.Query().Get("status")
	items, err := h.db.ListQuarantineItems(r.Context(), filter)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = make([]database.QuarantineItem, 0)
	}
	_ = json.NewEncoder(w).Encode(items)
}

