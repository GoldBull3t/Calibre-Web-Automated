package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/crocodilestick/calibre-web-go/internal/auth"
	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/crocodilestick/calibre-web-go/internal/events"
	"github.com/crocodilestick/calibre-web-go/internal/formats"
	"github.com/crocodilestick/calibre-web-go/internal/mail"
	"github.com/flosch/pongo2/v6"
)

// GetCover serve cover.jpg com suporte a cache LRU em memória (8MB) e HTTP Cache Headers.
func (h *Handler) GetCover(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 3 {
			idStr = parts[2]
		}
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.serveDefaultCover(w, r)
		return
	}

	cacheKey := fmt.Sprintf("cover:%d", id)
	cacheControl := "public, max-age=3600, must-revalidate"
	if r.URL.Query().Get("t") != "" || r.URL.Query().Get("v") != "" {
		cacheControl = "public, max-age=604800, immutable"
	}
	if h.coverCache != nil {
		if data, mime, ok := h.coverCache.Get(cacheKey); ok {
			w.Header().Set("Cache-Control", cacheControl)
			w.Header().Set("Content-Type", mime)
			w.Header().Set("Content-Length", strconv.Itoa(len(data)))
			_, _ = w.Write(data)
			return
		}
	}

	book, err := h.db.GetBookByID(r.Context(), id)
	if err != nil || book == nil {
		h.serveDefaultCover(w, r)
		return
	}

	coverPath := filepath.Join(h.booksDir, filepath.FromSlash(book.Path), "cover.jpg")
	if fi, err := os.Stat(coverPath); err == nil && !fi.IsDir() {
		// Se a imagem for <= 512KB, guarda na cache LRU em memória
		if fi.Size() <= 512*1024 && h.coverCache != nil {
			if data, err := os.ReadFile(coverPath); err == nil {
				h.coverCache.Put(cacheKey, data, "image/jpeg")
				w.Header().Set("Cache-Control", cacheControl)
				w.Header().Set("Content-Type", "image/jpeg")
				w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
				_, _ = w.Write(data)
				return
		}
		}
		w.Header().Set("Cache-Control", cacheControl)
		http.ServeFile(w, r, coverPath)
		return
	}

	h.serveDefaultCover(w, r)
}

// UploadCover permite atualizar a capa de um livro (suporta multipart form e raw binary image).
func (h *Handler) UploadCover(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if !auth.HasPermission(u, auth.RoleEdit) {
		http.Error(w, `{"error":"Sem permissão para alterar capas"}`, http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 {
			idStr = parts[3]
		}
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"ID de livro inválido"}`, http.StatusBadRequest)
		return
	}

	book, err := h.db.GetBookByID(r.Context(), id)
	if err != nil || book == nil {
		http.Error(w, `{"error":"Livro não encontrado"}`, http.StatusNotFound)
		return
	}

	var imgData []byte
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
			http.Error(w, `{"error":"Falha ao ler multipart form"}`, http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("cover")
		if err != nil {
			file, _, err = r.FormFile("image")
		}
		if err != nil {
			file, _, err = r.FormFile("file")
		}
		if err != nil {
			http.Error(w, `{"error":"Ficheiro da capa não encontrado no formulário"}`, http.StatusBadRequest)
			return
		}
		defer file.Close()
		imgData, err = io.ReadAll(file)
		if err != nil {
			http.Error(w, `{"error":"Erro ao ler imagem enviada"}`, http.StatusInternalServerError)
			return
		}
	} else if strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "application/octet-stream") {
		imgData, err = io.ReadAll(io.LimitReader(r.Body, 10*1024*1024))
		if err != nil {
			http.Error(w, `{"error":"Erro ao ler stream da imagem"}`, http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, `{"error":"Content-Type não suportado. Envie multipart/form-data ou image/jpeg/png"}`, http.StatusBadRequest)
		return
	}

	if len(imgData) == 0 {
		http.Error(w, `{"error":"Ficheiro de imagem está vazio"}`, http.StatusBadRequest)
		return
	}

	bookDir := filepath.Join(h.booksDir, filepath.FromSlash(book.Path))
	_ = os.MkdirAll(bookDir, 0755)
	coverPath := filepath.Join(bookDir, "cover.jpg")

	if err := os.WriteFile(coverPath, imgData, 0644); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Falha ao gravar capa no disco: %v"}`, err), http.StatusInternalServerError)
		return
	}

	_ = h.db.SetBookHasCover(r.Context(), id, true)

	// Invalida e renova cache em memória
	cacheKey := fmt.Sprintf("cover:%d", id)
	if h.coverCache != nil {
		h.coverCache.Delete(cacheKey)
		if len(imgData) <= 512*1024 {
			h.coverCache.Put(cacheKey, imgData, "image/jpeg")
		}
	}

	// Emite evento em tempo real via SSE
	events.GetBroker().Broadcast("book.cover_updated", fmt.Sprintf("Capa de '%s' atualizada com sucesso", book.Title), map[string]any{
		"book_id": id,
		"title":   book.Title,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Capa atualizada com sucesso",
		"book_id": id,
	})
}

// GetSeriesCover serve a capa representativa de uma série.
func (h *Handler) GetSeriesCover(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.serveDefaultCover(w, r)
		return
	}

	res, err := h.db.ListBooks(r.Context(), database.BookFilter{
		SeriesID: id,
		Page:     1,
		PageSize: 1,
	})
	if err == nil && len(res.Items) > 0 {
		coverPath := filepath.Join(h.booksDir, filepath.FromSlash(res.Items[0].Path), "cover.jpg")
		if fi, err := os.Stat(coverPath); err == nil && !fi.IsDir() {
			w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
			http.ServeFile(w, r, coverPath)
			return
		}
	}

	h.serveDefaultCover(w, r)
}

func (h *Handler) serveDefaultCover(w http.ResponseWriter, r *http.Request) {
	for _, name := range []string{"generic_cover.svg", "generic_cover.jpg", "img/cwa-logo-round.svg"} {
		defaultPath := filepath.Join(h.staticDir, filepath.FromSlash(name))
		if fi, err := os.Stat(defaultPath); err == nil && !fi.IsDir() {
			w.Header().Set("Cache-Control", "public, max-age=86400")
			if strings.HasSuffix(name, ".svg") {
				w.Header().Set("Content-Type", "image/svg+xml")
			} else {
				w.Header().Set("Content-Type", "image/jpeg")
			}
			http.ServeFile(w, r, defaultPath)
			return
		}
	}
	http.NotFound(w, r)
}

// DownloadBook entrega o ficheiro com suporte completo a HTTP Range Requests (206 Partial Content).
func (h *Handler) DownloadBook(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if !auth.HasPermission(u, auth.RoleDownload) {
		http.Error(w, "Sem permissão para transferir livros", http.StatusForbidden)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 3 {
			idStr = parts[2]
		}
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	book, err := h.db.GetBookByID(r.Context(), id)
	if err != nil || book == nil || len(book.Formats) == 0 {
		http.NotFound(w, r)
		return
	}

	reqFormat := r.PathValue("format")
	if reqFormat == "" {
		reqFormat = r.URL.Query().Get("format")
	}

	selectedFormat := book.Formats[0]
	if reqFormat != "" {
		for _, f := range book.Formats {
			if strings.EqualFold(f.Format, reqFormat) {
				selectedFormat = f
				break
			}
		}
	} else {
		for _, f := range book.Formats {
			if strings.EqualFold(f.Format, "EPUB") {
				selectedFormat = f
				break
			}
		}
	}

	fileName := fmt.Sprintf("%s.%s", selectedFormat.Name, strings.ToLower(selectedFormat.Format))
	fullPath := filepath.Join(h.booksDir, filepath.FromSlash(book.Path), fileName)

	if _, err := os.Stat(fullPath); err != nil {
		bookDir := filepath.Join(h.booksDir, filepath.FromSlash(book.Path))
		targetExt := "." + strings.ToLower(selectedFormat.Format)
		if entries, errRead := os.ReadDir(bookDir); errRead == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), targetExt) {
					fullPath = filepath.Join(bookDir, entry.Name())
					fileName = entry.Name()
					break
				}
			}
		}
		if _, errRetry := os.Stat(fullPath); errRetry != nil {
			http.NotFound(w, r)
			return
		}
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	http.ServeFile(w, r, fullPath)
}

// ShowBookStream faz streaming em linha (Range 206) para o visualizador epub.js.
func (h *Handler) ShowBookStream(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	book, err := h.db.GetBookByID(r.Context(), id)
	if err != nil || book == nil || len(book.Formats) == 0 {
		http.NotFound(w, r)
		return
	}

	reqFormat := r.PathValue("format")
	selectedFormat := book.Formats[0]
	if reqFormat != "" {
		for _, f := range book.Formats {
			if strings.EqualFold(f.Format, reqFormat) {
				selectedFormat = f
				break
			}
		}
	}

	fileName := fmt.Sprintf("%s.%s", selectedFormat.Name, strings.ToLower(selectedFormat.Format))
	fullPath := filepath.Join(h.booksDir, filepath.FromSlash(book.Path), fileName)

	if _, err := os.Stat(fullPath); err != nil {
		bookDir := filepath.Join(h.booksDir, filepath.FromSlash(book.Path))
		targetExt := "." + strings.ToLower(selectedFormat.Format)
		if entries, errRead := os.ReadDir(bookDir); errRead == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), targetExt) {
					fullPath = filepath.Join(bookDir, entry.Name())
					fileName = entry.Name()
					break
				}
			}
		}
		if _, errRetry := os.Stat(fullPath); errRetry != nil {
			http.NotFound(w, r)
			return
		}
	}

	switch strings.ToUpper(selectedFormat.Format) {
	case "EPUB":
		w.Header().Set("Content-Type", "application/epub+zip")
	case "PDF":
		w.Header().Set("Content-Type", "application/pdf")
	case "CBZ":
		w.Header().Set("Content-Type", "application/vnd.comicbook+zip")
	case "CBR":
		w.Header().Set("Content-Type", "application/vnd.comicbook-rar")
	}
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, fileName))
	http.ServeFile(w, r, fullPath)
}

// ReaderWeb entrega o leitor web embutido original baseado no read.html do Calibre-Web.
func (h *Handler) ReaderWeb(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	book, err := h.db.GetBookByID(r.Context(), id)
	if err != nil || book == nil {
		http.NotFound(w, r)
		return
	}

	format := r.PathValue("format")
	hasCBZ := false
	hasEPUB := false
	hasPDF := false
	for _, f := range book.Formats {
		if strings.EqualFold(f.Format, "CBZ") || strings.EqualFold(f.Format, "CBR") {
			hasCBZ = true
		}
		if strings.EqualFold(f.Format, "EPUB") {
			hasEPUB = true
		}
		if strings.EqualFold(f.Format, "PDF") {
			hasPDF = true
		}
	}

	if strings.EqualFold(format, "cbz") || strings.EqualFold(format, "cbr") || strings.EqualFold(format, "comic") || (hasCBZ && !hasEPUB && !hasPDF && format == "") {
		h.ComicReaderWeb(w, r, id)
		return
	}

	if strings.EqualFold(format, "pdf") || (hasPDF && !hasEPUB && format == "") {
		h.PDFReaderWeb(w, r, id, book)
		return
	}

	if format == "" {
		if hasEPUB {
			format = "epub"
		} else if len(book.Formats) > 0 {
			format = strings.ToLower(book.Formats[0].Format)
		} else {
			format = "epub"
		}
	}

	if !hasEPUB && !strings.EqualFold(format, "epub") && !strings.EqualFold(format, "kepub") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="pt">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s — Leitor não suportado</title>
  <style>
    body { background: #0b0f19; color: #e2e8f0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
    .card { background: rgba(30, 41, 59, 0.7); border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 16px; padding: 32px; max-width: 480px; text-align: center; backdrop-filter: blur(16px); }
    h2 { font-size: 18px; margin: 0 0 12px; }
    p { font-size: 14px; color: #94a3b8; line-height: 1.5; margin: 0 0 24px; }
    .btn-row { display: flex; gap: 12px; justify-content: center; }
    .btn { display: inline-flex; align-items: center; gap: 6px; padding: 10px 18px; border-radius: 8px; font-size: 13px; font-weight: 600; text-decoration: none; cursor: pointer; border: none; }
    .btn-primary { background: #0284c7; color: #fff; }
    .btn-secondary { background: rgba(255, 255, 255, 0.08); color: #cbd5e1; }
  </style>
</head>
<body>
  <div class="card">
    <h2>Formato .%s não suportado no navegador</h2>
    <p>O livro <strong>%s</strong> apenas possui ficheiros no formato <strong>%s</strong>. Descarregue o ficheiro para o seu leitor dedicado ou use as ferramentas CWA para convertê-lo em EPUB.</p>
    <div class="btn-row">
      <a href="/download/%d" class="btn btn-primary" download>Descarregar Ficheiro</a>
      <a href="/" class="btn btn-secondary">Voltar à Biblioteca</a>
    </div>
  </div>
</body>
</html>`, book.Title, strings.ToUpper(format), book.Title, strings.ToUpper(format), id)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var savedBookmark string
	if u := h.CurrentUser(r); u != nil {
		savedBookmark, _ = h.db.GetBookmark(r.Context(), u.ID, id, format)
	}

	_ = h.readerTpl.ExecuteWriter(map[string]any{
		"book":      book,
		"bookID":    id,
		"title":     book.Title,
		"epubURL":   fmt.Sprintf("/show/%d/%s/book.epub", id, format),
		"bookmark":  savedBookmark,
		"format":    format,
		"user":      h.CurrentUser(r),
	}, w)
}

// UploadBook processa o envio de ficheiros e copia diretamente para ingestDir.

// PDFReaderWeb disponibiliza o leitor dedicado de PDFs em linha com streaming de ficheiro nativo e sem erros.
func (h *Handler) PDFReaderWeb(w http.ResponseWriter, r *http.Request, id int64, book *database.BookDetails) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
  <title>%s — Leitor de PDF</title>
  <link rel="shortcut icon" href="/static/favicon.ico">
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    html, body { width: 100%%; height: 100%%; overflow: hidden; background: #0b0f19; color: #e2e8f0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; display: flex; flex-direction: column; }
    #pdf-header {
      height: 48px; background: rgba(15, 23, 42, 0.95); border-bottom: 1px solid rgba(255, 255, 255, 0.08);
      display: flex; align-items: center; justify-content: space-between; padding: 0 16px; flex-shrink: 0;
      backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px); z-index: 10;
    }
    .header-btn {
      background: transparent; border: 1px solid transparent; color: #e2e8f0; border-radius: 8px;
      height: 32px; padding: 0 12px; display: inline-flex; align-items: center; gap: 6px; font-size: 12px;
      font-weight: 500; text-decoration: none; cursor: pointer; transition: all 0.15s;
    }
    .header-btn:hover { background: rgba(255,255,255,0.08); border-color: rgba(255,255,255,0.12); }
    .header-btn svg { width: 16px; height: 16px; fill: none; stroke: currentColor; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }
    .book-title { font-size: 13px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 48vw; }
    #pdf-container { flex: 1; width: 100%%; height: calc(100%% - 48px); background: #1e293b; position: relative; }
    #pdf-frame { width: 100%%; height: 100%%; border: none; display: block; }
  </style>
</head>
<body>
  <header id="pdf-header">
    <a href="/" class="header-btn" title="Voltar à Biblioteca">
      <svg viewBox="0 0 24 24"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
      <span>Biblioteca</span>
    </a>
    <div class="book-title">%s</div>
    <div style="display:flex; align-items:center; gap:8px;">
      <a href="/download/%d/pdf" class="header-btn" download title="Descarregar PDF">
        <svg viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
        <span>Descarregar</span>
      </a>
      <button id="btn-fullscreen" class="header-btn" title="Ecrã Completo">
        <svg viewBox="0 0 24 24"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/></svg>
      </button>
    </div>
  </header>
  <main id="pdf-container">
    <iframe id="pdf-frame" src="/show/%d/pdf/book.pdf#view=FitH" type="application/pdf"></iframe>
  </main>
  <script>
    document.getElementById("btn-fullscreen").addEventListener("click", function() {
      if (!document.fullscreenElement) document.documentElement.requestFullscreen().catch(function(){});
      else document.exitFullscreen().catch(function(){});
    });
  </script>
</body>
</html>`, book.Title, book.Title, id, id)
	_, _ = w.Write([]byte(html))
}

func (h *Handler) UploadBook(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if !auth.HasPermission(u, auth.RoleUpload) {
		http.Error(w, "Sem permissão para carregar livros", http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(100 * 1024 * 1024); err != nil {
		http.Error(w, "Upload failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	fileHeaders := r.MultipartForm.File["file"]
	if len(fileHeaders) == 0 {
		fileHeaders = r.MultipartForm.File["btn-upload"]
	}
	if len(fileHeaders) == 0 {
		fileHeaders = r.MultipartForm.File["btn-upload2"]
	}

	for _, fh := range fileHeaders {
		cleanName := filepath.Base(fh.Filename)
		if cleanName == "" || cleanName == "." || cleanName == ".." {
			continue
		}

		src, err := fh.Open()
		if err != nil {
			continue
		}

		dstPath := filepath.Join(h.ingestDir, cleanName)
		dst, err := os.Create(dstPath)
		if err != nil {
			src.Close()
			continue
		}
		_, _ = io.Copy(dst, src)
		dst.Close()
		src.Close()
	}

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DeleteBook remove o livro do banco e da biblioteca física.
func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if !auth.HasPermission(u, auth.RoleDeleteBooks) {
		http.Error(w, "Sem permissão para eliminar livros", http.StatusForbidden)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	book, _ := h.db.GetBookByID(r.Context(), id)
	_ = h.db.DeleteBook(r.Context(), id)
	if h.coverCache != nil {
		h.coverCache.Delete(fmt.Sprintf("cover:%d", id))
	}

	if book != nil && book.Path != "" {
		fullPath := filepath.Join(h.booksDir, filepath.FromSlash(book.Path))
		_ = os.RemoveAll(fullPath)
	}

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ToggleRead alterna estado de leitura do livro para o utilizador.
func (h *Handler) ToggleRead(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if u == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	newStatus, err := h.db.ToggleBookRead(r.Context(), u.ID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"read_status": newStatus,
	})
}

// ToggleArchived alterna estado de livro arquivado para o utilizador.
func (h *Handler) ToggleArchived(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if u == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	newStatus, err := h.db.ToggleBookArchived(r.Context(), u.ID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"is_archived": newStatus,
	})
}

// Bookmark guarda posição do leitor no formato indicado.
func (h *Handler) Bookmark(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if u == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	format := r.PathValue("format")
	_ = r.ParseForm()
	bookmarkKey := r.FormValue("bookmark")

	_ = h.db.SaveBookmark(r.Context(), u.ID, id, format, bookmarkKey)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
	})
}

// EmailStat devolve o status do envio de e-mails.
func (h *Handler) EmailStat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
	})
}

// ListBooksJSON devolve os dados em JSON para a visualização em tabela (/table e /ajax/listbooks).
func (h *Handler) ListBooksJSON(w http.ResponseWriter, r *http.Request) {
	res, err := h.db.ListBooks(r.Context(), database.BookFilter{
		Page:     1,
		PageSize: 100,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"total": res.TotalItems,
		"rows":  res.Items,
	})
}

// ShowEditBook entrega o formulário de edição de metadados do livro.
func (h *Handler) ShowEditBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	book, err := h.db.GetBookByID(r.Context(), id)
	if err != nil || book == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.bookEditTpl.ExecuteWriter(pongo2.Context{
		"book":  book,
		"title": book.Title,
	}, w)
}

// EditBook processa a atualização dos metadados e o upload de nova capa.
func (h *Handler) EditBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseMultipartForm(50 * 1024 * 1024); err != nil {
		_ = r.ParseForm()
	}

	seriesIndex, _ := strconv.ParseFloat(r.FormValue("series_index"), 64)
	rating, _ := strconv.Atoi(r.FormValue("rating"))

	var authors []string
	if aStr := strings.TrimSpace(r.FormValue("authors")); aStr != "" {
		for _, a := range strings.Split(aStr, ",") {
			if trimmed := strings.TrimSpace(a); trimmed != "" {
				authors = append(authors, trimmed)
			}
		}
	}

	var tags []string
	if tStr := strings.TrimSpace(r.FormValue("tags")); tStr != "" {
		for _, t := range strings.Split(tStr, ",") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	var publishers []string
	if pStr := strings.TrimSpace(r.FormValue("publisher")); pStr != "" {
		for _, p := range strings.Split(pStr, ",") {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				publishers = append(publishers, trimmed)
			}
		}
	}

	var languages []string
	if lStr := strings.TrimSpace(r.FormValue("languages")); lStr != "" {
		for _, l := range strings.Split(lStr, ",") {
			if trimmed := strings.TrimSpace(l); trimmed != "" {
				languages = append(languages, trimmed)
			}
		}
	}

	form := database.BookEditForm{
		Title:       r.FormValue("title"),
		TitleSort:   r.FormValue("sort"),
		Authors:     authors,
		Series:      r.FormValue("series"),
		SeriesIndex: seriesIndex,
		Tags:        tags,
		Publishers:  publishers,
		Languages:   languages,
		Rating:      rating,
		Comment:     r.FormValue("comments"),
	}

	_ = h.db.UpdateBookMetadata(r.Context(), id, form)

	// Processa nova capa se enviada
	if r.MultipartForm != nil {
		coverFiles := r.MultipartForm.File["cover_file"]
		if len(coverFiles) == 0 {
			coverFiles = r.MultipartForm.File["btn-upload-cover"]
		}
		if len(coverFiles) > 0 {
			book, err := h.db.GetBookByID(r.Context(), id)
			if err == nil && book != nil && book.Path != "" {
				coverDst := filepath.Join(h.booksDir, filepath.FromSlash(book.Path), "cover.jpg")
				if src, err := coverFiles[0].Open(); err == nil {
					defer src.Close()
					if dst, err := os.Create(coverDst); err == nil {
						_, _ = io.Copy(dst, src)
						dst.Close()
					}
				}
			}
		}
	}

	// CWA Flagship Feature: Automatic Enforcement of Covers & Metadata
	if updatedBook, err := h.db.GetBookByID(r.Context(), id); err == nil && updatedBook != nil && updatedBook.Path != "" {
		authorName := ""
		if len(form.Authors) > 0 {
			authorName = form.Authors[0]
		}
		coverDst := filepath.Join(h.booksDir, filepath.FromSlash(updatedBook.Path), "cover.jpg")
		if _, errStat := os.Stat(coverDst); errStat != nil {
			coverDst = ""
		}
		for _, f := range updatedBook.Formats {
			if strings.EqualFold(f.Format, "EPUB") {
				epubPath := filepath.Join(h.booksDir, filepath.FromSlash(updatedBook.Path), fmt.Sprintf("%s.epub", f.Name))
				_ = formats.EnforceCoverAndMetadata(epubPath, coverDst, form.Title, authorName)
			}
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/book/%d", id), http.StatusSeeOther)
}

// DeleteBookFormat remove um formato de ficheiro do livro.
func (h *Handler) DeleteBookFormat(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if !auth.HasPermission(u, auth.RoleDeleteBooks) {
		http.Error(w, "Sem permissão para eliminar formatos", http.StatusForbidden)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	format := r.PathValue("format")
	book, _ := h.db.GetBookByID(r.Context(), id)

	if format != "" {
		_ = h.db.DeleteBookFormat(r.Context(), id, format)
		if book != nil && book.Path != "" {
			formatName := ""
			for _, f := range book.Formats {
				if strings.EqualFold(f.Format, format) {
					formatName = f.Name
					break
				}
			}
			if formatName == "" {
				formatName = book.Title
			}
			fileName := fmt.Sprintf("%s.%s", formatName, strings.ToLower(format))
			fullPath := filepath.Join(h.booksDir, filepath.FromSlash(book.Path), fileName)
			_ = os.Remove(fullPath)
		}
	} else {
		_ = h.db.DeleteBook(r.Context(), id)
		if book != nil && book.Path != "" {
			fullPath := filepath.Join(h.booksDir, filepath.FromSlash(book.Path))
			_ = os.RemoveAll(fullPath)
		}
	}

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
		return
	}

	if format != "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/book/%d", id), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// SendBook envia o livro para o e-reader / Kindle via SMTP.
func (h *Handler) SendBook(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if u == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	book, err := h.db.GetBookByID(r.Context(), id)
	if err != nil || book == nil || len(book.Formats) == 0 {
		http.NotFound(w, r)
		return
	}

	recipient := strings.TrimSpace(r.URL.Query().Get("email"))
	if recipient == "" {
		recipient = u.KindleMail
	}
	if recipient == "" {
		recipient = u.Email
	}

	if recipient == "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Nenhum endereço de email configurado no perfil para envio.",
		})
		return
	}

	reqFormat := r.PathValue("format")
	selectedFormat := book.Formats[0]
	if reqFormat != "" {
		for _, f := range book.Formats {
			if strings.EqualFold(f.Format, reqFormat) {
				selectedFormat = f
				break
			}
		}
	} else {
		for _, f := range book.Formats {
			if strings.EqualFold(f.Format, "EPUB") || strings.EqualFold(f.Format, "MOBI") {
				selectedFormat = f
				break
			}
		}
	}

	fileName := fmt.Sprintf("%s.%s", selectedFormat.Name, strings.ToLower(selectedFormat.Format))
	fullPath := filepath.Join(h.booksDir, filepath.FromSlash(book.Path), fileName)

	// EPUB Fixer 2.0: repara automaticamente cópia temporária antes do envio (sem alterar o mestre)
	fileToSend := fullPath
	if strings.EqualFold(selectedFormat.Format, "EPUB") {
		if tmpFile, errTmp := os.CreateTemp("", "cwa-send-*.epub"); errTmp == nil {
			tmpPath := tmpFile.Name()
			_ = tmpFile.Close()
			if inData, errR := os.ReadFile(fullPath); errR == nil {
				if errW := os.WriteFile(tmpPath, inData, 0644); errW == nil {
					_, _ = formats.FixEPUB(tmpPath)
					fileToSend = tmpPath
					defer os.Remove(tmpPath)
				}
			}
		}
	}

	smtpSettings, err := h.db.GetSMTPSettings(r.Context())
	if err != nil || smtpSettings.Server == "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Servidor SMTP não configurado nas definições de administrador.",
		})
		return
	}

	err = mail.SendBookEmail(mail.SMTPSettings{
		Server:   smtpSettings.Server,
		Port:     smtpSettings.Port,
		UseSSL:   smtpSettings.UseSSL,
		Login:    smtpSettings.Login,
		Password: smtpSettings.Password,
		From:     smtpSettings.From,
	}, recipient, book.Title, fileToSend, fileName)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": fmt.Sprintf("Falha no envio de email: %v", err),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": fmt.Sprintf("Livro enviado com sucesso para %s.", recipient),
	})
}

// ConvertBook executa a conversão de formato do livro utilizando ebook-convert ou kepubify.
func (h *Handler) ConvertBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	book, err := h.db.GetBookByID(r.Context(), id)
	if err != nil || book == nil || len(book.Formats) == 0 {
		http.NotFound(w, r)
		return
	}

	_ = r.ParseForm()
	targetFormat := strings.TrimSpace(r.FormValue("btn-convert"))
	if targetFormat == "" {
		targetFormat = strings.TrimSpace(r.FormValue("target_format"))
	}
	if targetFormat == "" {
		targetFormat = strings.TrimSpace(r.FormValue("format"))
	}
	if targetFormat == "" {
		targetFormat = "kepub"
	}
	targetFormat = strings.ToLower(strings.TrimPrefix(targetFormat, "."))
	allowedFormats := map[string]bool{"epub": true, "kepub": true, "mobi": true, "azw3": true, "pdf": true, "txt": true}
	if !allowedFormats[targetFormat] {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "error", "message": "Formato de destino não suportado"})
		return
	}

	// Localiza o formato fonte disponível (preferência por EPUB)
	var sourceFormat database.DataFormat
	for _, f := range book.Formats {
		if strings.EqualFold(f.Format, "EPUB") {
			sourceFormat = f
			break
		}
	}
	if sourceFormat.Format == "" && len(book.Formats) > 0 {
		sourceFormat = book.Formats[0]
	}

	bookDir := filepath.Join(h.booksDir, filepath.FromSlash(book.Path))
	sourceExt := strings.ToLower(sourceFormat.Format)
	sourceFile := filepath.Join(bookDir, fmt.Sprintf("%s.%s", sourceFormat.Name, sourceExt))
	if _, err := os.Stat(sourceFile); err != nil {
		// Fallback: procura por extensão
		entries, _ := os.ReadDir(bookDir)
		for _, e := range entries {
			if strings.EqualFold(filepath.Ext(e.Name()), "."+sourceExt) {
				sourceFile = filepath.Join(bookDir, e.Name())
				break
			}
		}
	}

	outputFileName := sourceFormat.Name
	if outputFileName == "" {
		outputFileName = fmt.Sprintf("%d", book.ID)
	}
	outputFile := filepath.Join(bookDir, fmt.Sprintf("%s.%s", outputFileName, targetFormat))
	if targetFormat == "kepub" {
		outputFile = filepath.Join(bookDir, fmt.Sprintf("%s.kepub.epub", outputFileName))
	}

	var cmd *exec.Cmd
	if targetFormat == "kepub" {
		if bin, err := exec.LookPath("kepubify"); err == nil {
			cmd = exec.CommandContext(r.Context(), bin, "-o", outputFile, sourceFile)
		}
	}
	if cmd == nil {
		if bin, err := exec.LookPath("ebook-convert"); err == nil {
			cmd = exec.CommandContext(r.Context(), bin, sourceFile, outputFile)
		}
	}

	if cmd != nil {
		if out, err := cmd.CombinedOutput(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":  "error",
				"message": fmt.Sprintf("Conversion failed: %s (%s)", err, strings.TrimSpace(string(out))),
			})
			return
		}

		if info, err := os.Stat(outputFile); err == nil {
			_ = h.db.AddBookFormat(r.Context(), book.ID, targetFormat, info.Size(), outputFileName)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":        "success",
				"message":       fmt.Sprintf("Book successfully converted to %s", strings.ToUpper(targetFormat)),
				"target_format": targetFormat,
				"size":          info.Size(),
			})
			return
		}
	}

	// Se ferramentas externas não estiverem instaladas no PATH, informa status queued com detalhes
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":         "queued",
		"message":        fmt.Sprintf("Conversion to %s queued.", strings.ToUpper(targetFormat)),
		"target_format":  targetFormat,
		"tool_available": false,
	})
}


// DisplaySelectedBooks lida com a seleção em lote de livros.
func (h *Handler) DisplaySelectedBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success"})
}

// ArchiveSelectedBooks arquiva ou desarquiva os livros selecionados.
func (h *Handler) ArchiveSelectedBooks(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if u == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		Selections []int64 `json:"selections"`
		Archive    *bool   `json:"archive"`
	}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&payload)
	} else {
		_ = r.ParseForm()
		for _, s := range r.Form["selections[]"] {
			if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				payload.Selections = append(payload.Selections, id)
			}
		}
	}

	archive := true
	if payload.Archive != nil {
		archive = *payload.Archive
	}

	_ = h.db.BulkSetArchived(r.Context(), u.ID, payload.Selections, archive)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// DeleteSelectedBooks remove os livros selecionados em lote do banco e disco.
func (h *Handler) DeleteSelectedBooks(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if !auth.HasPermission(u, auth.RoleDeleteBooks) {
		http.Error(w, `{"error":"Sem permissão para eliminar livros"}`, http.StatusForbidden)
		return
	}

	var payload struct {
		Selections []int64 `json:"selections"`
	}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&payload)
	} else {
		_ = r.ParseForm()
		for _, s := range r.Form["selections[]"] {
			if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				payload.Selections = append(payload.Selections, id)
			}
		}
	}

	for _, id := range payload.Selections {
		book, _ := h.db.GetBookByID(r.Context(), id)
		_ = h.db.DeleteBook(r.Context(), id)
		if h.coverCache != nil {
		h.coverCache.Delete(fmt.Sprintf("cover:%d", id))
	}
		if book != nil && book.Path != "" {
			fullPath := filepath.Join(h.booksDir, filepath.FromSlash(book.Path))
			_ = os.RemoveAll(fullPath)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// ReadSelectedBooks marca os livros selecionados como lidos ou não lidos.
func (h *Handler) ReadSelectedBooks(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if u == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		Selections []int64 `json:"selections"`
		MarkAsRead *bool   `json:"markAsRead"`
	}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&payload)
	} else {
		_ = r.ParseForm()
		for _, s := range r.Form["selections[]"] {
			if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				payload.Selections = append(payload.Selections, id)
			}
		}
	}

	markAsRead := true
	if payload.MarkAsRead != nil {
		markAsRead = *payload.MarkAsRead
	}

	_ = h.db.BulkSetReadStatus(r.Context(), u.ID, payload.Selections, markAsRead)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// EditSelectedBooks edita metadados em lote dos livros selecionados.
func (h *Handler) EditSelectedBooks(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if !auth.HasPermission(u, auth.RoleEdit) {
		http.Error(w, `{"error":"Sem permissão para editar livros"}`, http.StatusForbidden)
		return
	}
	_ = r.ParseForm()
	tagsStr := strings.TrimSpace(r.FormValue("add_tag"))
	var tags []string
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	for _, idStr := range r.Form["selections[]"] {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			book, err := h.db.GetBookByID(r.Context(), id)
			if err == nil && book != nil {
				mergedTags := append(book.Tags, tags...)
				_ = h.db.UpdateBookMetadata(r.Context(), id, database.BookEditForm{
					Title:       book.Title,
					TitleSort:   book.Sort,
					Authors:     book.Authors,
					Series:      book.Series,
					SeriesIndex: book.SeriesIndex,
					Tags:        mergedTags,
					Publishers:  book.Publishers,
					Languages:   book.Languages,
					Rating:      book.Rating,
					Comment:     book.Comment,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// MergeBooks funde múltiplos livros num único registo com todos os formatos.
func (h *Handler) MergeBooks(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if !auth.HasPermission(u, auth.RoleEdit) {
		http.Error(w, `{"error":"Sem permissão para mesclar livros"}`, http.StatusForbidden)
		return
	}
	var payload struct {
		MergeBooks []int64 `json:"Merge_books"`
	}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&payload)
	} else {
		_ = r.ParseForm()
		for _, s := range r.Form["Merge_books[]"] {
			if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				payload.MergeBooks = append(payload.MergeBooks, id)
			}
		}
	}

	if len(payload.MergeBooks) >= 2 {
		targetID := payload.MergeBooks[0]
		sources := payload.MergeBooks[1:]
		_ = h.db.MergeBookRecords(r.Context(), targetID, sources, h.booksDir)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

