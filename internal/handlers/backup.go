package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/crocodilestick/calibre-web-go/internal/backup"
	"github.com/flosch/pongo2/v6"
)

// CWABackupList devolve a lista de backups disponíveis em HTML ou JSON.
func (h *Handler) CWABackupList(w http.ResponseWriter, r *http.Request) {
	backupDir := filepath.Join(h.libraryDir, "backups")
	backups, err := backup.ListBackups(backupDir)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": %q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("format") == "json" || (strings.Contains(r.Header.Get("Accept"), "application/json") && !strings.Contains(r.Header.Get("Accept"), "text/html")) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "success",
			"backups": backups,
		})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.cwaBackupTpl.ExecuteWriter(pongo2.Context{"backups": backups}, w)
}

// CWABackupCreate aciona a criação imediata de um snapshot de backup compactado (.zip).
func (h *Handler) CWABackupCreate(w http.ResponseWriter, r *http.Request) {
	backupDir := filepath.Join(h.libraryDir, "backups")
	name, err := backup.CreateBackup(h.libraryDir, backupDir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "success",
		"filename": name,
		"message":  "Backup criado e comprimido com sucesso.",
	})
}

// CWABackupDownload transfere o ficheiro de backup por streaming.
func (h *Handler) CWABackupDownload(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		filename = filepath.Base(r.URL.Path)
	}

	// Higienização contra Path Traversal
	cleanName := filepath.Base(filename)
	if !strings.HasSuffix(strings.ToLower(cleanName), ".zip") {
		http.Error(w, "Ficheiro inválido", http.StatusBadRequest)
		return
	}

	backupPath := filepath.Join(h.libraryDir, "backups", cleanName)
	if _, err := os.Stat(backupPath); err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, cleanName))
	http.ServeFile(w, r, backupPath)
}

// CWABackupDelete apaga um backup específico.
func (h *Handler) CWABackupDelete(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	cleanName := filepath.Base(filename)
	backupPath := filepath.Join(h.libraryDir, "backups", cleanName)

	if err := os.Remove(backupPath); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Backup eliminado com sucesso.",
	})
}

