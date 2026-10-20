package spa

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var distFS embed.FS

// Handler fornece os assets estáticos compilados do frontend Svelte e gere o roteamento SPA.
type Handler struct {
	fileServer http.Handler
	indexHTML  []byte
	hasAssets  bool
}

// New inicializa o servidor de ficheiros embutidos do Svelte.
func New() (*Handler, error) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return &Handler{hasAssets: false}, nil
	}

	index, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		return &Handler{hasAssets: false}, nil
	}

	return &Handler{
		fileServer: http.FileServer(http.FS(sub)),
		indexHTML:  index,
		hasAssets:  true,
	}, nil
}

// HasAssets indica se os assets do Svelte estão compilados e prontos.
func (h *Handler) HasAssets() bool {
	return h != nil && h.hasAssets
}

// ServeHTTP serve a página inicial SPA ou os ficheiros estáticos /assets/*, sw.js, manifest.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.hasAssets {
		http.NotFound(w, r)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "sw.js" {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Service-Worker-Allowed", "/")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		h.fileServer.ServeHTTP(w, r)
		return
	}
	if path == "manifest.webmanifest" || path == "manifest.json" {
		w.Header().Set("Content-Type", "application/manifest+json; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		h.fileServer.ServeHTTP(w, r)
		return
	}
	if path == "favicon.svg" {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=604800")
		h.fileServer.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(path, "assets/") || path == "icons.svg" || strings.HasPrefix(path, "icons/") {
		h.fileServer.ServeHTTP(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(h.indexHTML)
}

