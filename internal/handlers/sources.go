package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/crocodilestick/calibre-web-go/internal/auth"
)

func jsonErr(w http.ResponseWriter, msg string, code int) {
	http.Error(w, `{"error":`+strconv.Quote(msg)+`}`, code)
}

// SourcesList devolve as fontes de livros e o seu estado (segredos mascarados). Só administradores.
func (h *Handler) SourcesList(w http.ResponseWriter, r *http.Request) {
	if !auth.HasPermission(h.CurrentUser(r), auth.RoleAdmin) {
		jsonErr(w, "Acesso reservado ao Administrador", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"sources": h.sourcesMgr.List()})
}

// SourcesConfig guarda a configuração (chaves/URLs) de uma fonte. Só administradores.
func (h *Handler) SourcesConfig(w http.ResponseWriter, r *http.Request) {
	if !auth.HasPermission(h.CurrentUser(r), auth.RoleAdmin) {
		jsonErr(w, "Acesso reservado ao Administrador", http.StatusForbidden)
		return
	}
	var req struct {
		Values  map[string]string `json:"values"`
		Enabled *bool             `json:"enabled"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		jsonErr(w, "pedido inválido", http.StatusBadRequest)
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if err := h.sourcesMgr.Configure(r.PathValue("id"), req.Values, enabled); err != nil {
		jsonErr(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "sources": h.sourcesMgr.List()})
}

// SourcesSearch pesquisa em todas as fontes ativas (ou numa só, com ?source=).
func (h *Handler) SourcesSearch(w http.ResponseWriter, r *http.Request) {
	if !auth.HasPermission(h.CurrentUser(r), auth.RoleDownload) {
		jsonErr(w, "Sem permissão de download", http.StatusForbidden)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < 2 {
		jsonErr(w, "pesquisa demasiado curta", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.sourcesMgr.Search(r.Context(), q, r.URL.Query().Get("source")))
}

// SourcesDownload descarrega um livro de uma fonte para a pasta de ingestão.
func (h *Handler) SourcesDownload(w http.ResponseWriter, r *http.Request) {
	if !auth.HasPermission(h.CurrentUser(r), auth.RoleDownload) {
		jsonErr(w, "Sem permissão de download", http.StatusForbidden)
		return
	}
	var req struct {
		Source string `json:"source"`
		ID     string `json:"id"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil || req.Source == "" || req.ID == "" {
		jsonErr(w, "source e id são obrigatórios", http.StatusBadRequest)
		return
	}
	path, err := h.sourcesMgr.DownloadToDir(r.Context(), req.Source, req.ID, h.ingestDir)
	if err != nil {
		jsonErr(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Livro descarregado para a pasta de ingestão.",
		"file":    path,
	})
}
