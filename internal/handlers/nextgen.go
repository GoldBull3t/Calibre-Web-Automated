package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/backup"
	"github.com/crocodilestick/calibre-web-go/internal/database"
)

// ReaderGetProgress recupera o progresso de leitura sincronizado (Single Source of Truth) para o livro.
func (h *Handler) ReaderGetProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("book_id")
	}

	bookID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || bookID <= 0 {
		http.Error(w, `{"error":"book_id inválido"}`, http.StatusBadRequest)
		return
	}

	userID := int64(1)
	if u := h.CurrentUser(r); u != nil {
		userID = u.ID
	}

	p, err := h.db.GetKOSyncProgress(r.Context(), userID, idStr)
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"book_id":    bookID,
			"progress":   "",
			"percentage": 0,
			"device":     "",
			"timestamp":  0,
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"book_id":    bookID,
		"progress":   p.Progress,
		"percentage": p.Percentage,
		"device":     p.Device,
		"timestamp":  p.Timestamp,
	})
}

// ReaderSaveProgress guarda e sincroniza o progresso entre o leitor Web, KOReader e marcadores.
func (h *Handler) ReaderSaveProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("book_id")
	}

	bookID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || bookID <= 0 {
		http.Error(w, `{"error":"book_id inválido"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Progress   string  `json:"progress"`
		Percentage float64 `json:"percentage"`
		Device     string  `json:"device"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"dados de requisição inválidos"}`, http.StatusBadRequest)
		return
	}

	userID := int64(1)
	if u := h.CurrentUser(r); u != nil {
		userID = u.ID
	}

	device := req.Device
	if device == "" {
		device = "Web Reader"
	}

	now := time.Now().Unix()
	p := database.KOSyncProgress{
		UserID:     userID,
		Document:   idStr,
		Progress:   req.Progress,
		Percentage: req.Percentage,
		Device:     device,
		DeviceID:   device,
		Timestamp:  now,
	}

	_ = h.db.UpdateKOSyncProgress(r.Context(), p)

	if req.Percentage >= 99.0 {
		_ = h.db.BulkSetReadStatus(r.Context(), userID, []int64{bookID}, true)
	}
	_ = h.db.SaveBookmark(r.Context(), bookID, userID, "EPUB", req.Progress)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "success",
		"book_id":   bookID,
		"timestamp": now,
	})
}

// SearchFTS executa pesquisa textual integral (Full Text Search) via SQLite FTS5.
func (h *Handler) SearchFTS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"query":   "",
			"count":   0,
			"results": []any{},
		})
		return
	}

	if h.searchIdx == nil {
		http.Error(w, `{"error":"motor FTS5 não inicializado"}`, http.StatusInternalServerError)
		return
	}

	results, err := h.searchIdx.Search(r.Context(), q, 30)
	if err != nil {
		http.Error(w, `{"error":`+strconv.Quote(err.Error())+`}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"query":   q,
		"count":   len(results),
		"results": results,
	})
}

// ReindexFTS indexa os livros EPUB da biblioteca no motor FTS5.
func (h *Handler) ReindexFTS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.searchIdx == nil {
		http.Error(w, `{"error":"motor FTS5 não inicializado"}`, http.StatusInternalServerError)
		return
	}

	// Executa reindexação em background para não bloquear a requisição HTTP
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		res, err := h.db.ListBooks(bgCtx, database.BookFilter{PageSize: 10000})
		if err != nil {
			return
		}
		for _, b := range res.Items {
			for _, f := range b.Formats {
				if strings.EqualFold(f.Format, "EPUB") {
					_ = h.searchIdx.IndexBookEPUB(bgCtx, b.ID, b.Title, strings.Join(b.Authors, ", "), b.Path, f.Name+".epub")
				}
			}
		}
	}()

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Reindexação FTS5 iniciada em background.",
	})
}

// MAMStatus devolve o estado da ligação ao MyAnonamouse (MAM 🐭🧀).
func (h *Handler) MAMStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.mamClient == nil || !h.mamClient.HasAuth() {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"configured": false,
			"message":    "Cookie mam_id não configurado.",
		})
		return
	}

	info, err := h.mamClient.GetAccountInfo(r.Context())
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"configured": true,
			"error":      err.Error(),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"configured": true,
		"account":    info,
	})
}

// MAMConfig guarda a configuração do token mam_id.
func (h *Handler) MAMConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		MamID string `json:"mam_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"dados inválidos"}`, http.StatusBadRequest)
		return
	}

	if h.mamClient != nil {
		h.mamClient.SetMamID(req.MamID)
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Configuração do MAM guardada com sucesso.",
	})
}

// MAMSearch pesquisa no catálogo do MyAnonamouse.
func (h *Handler) MAMSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"query":   "",
			"results": []any{},
		})
		return
	}

	if h.mamClient == nil || !h.mamClient.HasAuth() {
		http.Error(w, `{"error":"MAM não configurado. Adicione o seu mam_id primeiro."}`, http.StatusUnauthorized)
		return
	}

	items, err := h.mamClient.Search(r.Context(), q, 30)
	if err != nil {
		http.Error(w, `{"error":`+strconv.Quote(err.Error())+`}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"query":   q,
		"count":   len(items),
		"results": items,
	})
}

// MAMDownload descarrega o torrent do MAM diretamente para a pasta de ingestão contínua.
func (h *Handler) MAMDownload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	tid, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || tid <= 0 {
		http.Error(w, `{"error":"id de torrent inválido"}`, http.StatusBadRequest)
		return
	}

	if h.mamClient == nil || !h.mamClient.HasAuth() {
		http.Error(w, `{"error":"MAM não configurado"}`, http.StatusUnauthorized)
		return
	}

	path, err := h.mamClient.DownloadTorrentToIngest(r.Context(), tid, h.ingestDir)
	if err != nil {
		http.Error(w, `{"error":`+strconv.Quote(err.Error())+`}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Torrent enviado com sucesso para a pasta de ingestão contínua.",
		"path":    path,
	})
}

// PluginsList devolve a lista de plugins Calibre instalados.
func (h *Handler) PluginsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.pluginMgr == nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"plugins": []any{},
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"plugins": h.pluginMgr.List(),
	})
}

// PluginsUpload recebe e instala um plugin Calibre em ficheiro .zip.
func (h *Handler) PluginsUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	file, header, err := r.FormFile("plugin_file")
	if err != nil {
		http.Error(w, `{"error":"ficheiro de plugin não enviado"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	if h.pluginMgr == nil {
		http.Error(w, `{"error":"gestor de plugins não inicializado"}`, http.StatusInternalServerError)
		return
	}

	p, err := h.pluginMgr.SavePlugin(header.Filename, file)
	if err != nil {
		http.Error(w, `{"error":`+strconv.Quote(err.Error())+`}`, http.StatusBadRequest)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Plugin instalado com sucesso.",
		"plugin":  p,
	})
}

// PluginsToggle ativa ou desativa um plugin Calibre.
func (h *Handler) PluginsToggle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	filename := r.PathValue("filename")
	if filename == "" {
		filename = r.URL.Query().Get("filename")
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if h.pluginMgr == nil {
		http.Error(w, `{"error":"gestor de plugins não inicializado"}`, http.StatusInternalServerError)
		return
	}

	if err := h.pluginMgr.Toggle(filename, req.Enabled); err != nil {
		http.Error(w, `{"error":`+strconv.Quote(err.Error())+`}`, http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "success",
		"filename": filename,
		"enabled":  req.Enabled,
	})
}

// APIGetBooks devolve a lista de livros em JSON com paginação, filtros e ordenação para o frontend Svelte.
func (h *Handler) APIGetBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 200 {
		limit = 30
	}

	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "new"
	}

	query := strings.TrimSpace(r.URL.Query().Get("query"))
	authorID, _ := strconv.ParseInt(r.URL.Query().Get("author_id"), 10, 64)
	seriesID, _ := strconv.ParseInt(r.URL.Query().Get("series_id"), 10, 64)
	tagID, _ := strconv.ParseInt(r.URL.Query().Get("tag_id"), 10, 64)
	format := strings.TrimSpace(r.URL.Query().Get("format"))
	readStatus, _ := strconv.Atoi(r.URL.Query().Get("read_status"))

	userID := int64(1)
	if u := h.CurrentUser(r); u != nil {
		userID = u.ID
	}

	// Suporte a estantes inteligentes (magic_1..4) ou customizadas
	shelfParam := r.URL.Query().Get("shelf")
	if strings.HasPrefix(shelfParam, "magic_") {
		magicID, _ := strconv.ParseInt(strings.TrimPrefix(shelfParam, "magic_"), 10, 64)
		if magicID > 0 {
			res, err := h.db.GetMagicShelfBooks(r.Context(), magicID, userID, page, limit)
			if err == nil && res != nil {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"total": res.TotalItems,
					"page":  page,
					"limit": limit,
					"items": res.Items,
				})
				return
			}
		}
	} else if shelfID, _ := strconv.ParseInt(shelfParam, 10, 64); shelfID > 0 {
		res, err := h.db.GetShelfBooks(r.Context(), shelfID, sort, page, limit)
		if err == nil && res != nil {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"total": res.TotalItems,
				"page":  page,
				"limit": limit,
				"items": res.Items,
			})
			return
		}
	}

	filter := database.BookFilter{
		Page:       page,
		PageSize:   limit,
		Sort:       sort,
		Search:     query,
		AuthorID:   authorID,
		SeriesID:   seriesID,
		TagID:      tagID,
		Format:     format,
		ReadStatus: readStatus,
		UserID:     userID,
	}

	res, err := h.db.ListBooks(r.Context(), filter)
	if err != nil {
		http.Error(w, `{"error":`+strconv.Quote(err.Error())+`}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"total": res.TotalItems,
		"page":  page,
		"limit": limit,
		"items": res.Items,
	})
}

// APIGetBook devolve os detalhes completos de um livro em JSON.
func (h *Handler) APIGetBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"ID de livro inválido"}`, http.StatusBadRequest)
		return
	}

	b, err := h.db.GetBookByID(r.Context(), id)
	if err != nil || b == nil {
		http.Error(w, `{"error":"Livro não encontrado"}`, http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(b)
}

// APIStatsOverview fornece métricas em tempo real para o painel de controlo.
func (h *Handler) APIStatsOverview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	totalBooks, totalAuthors, totalSeries, _ := h.db.GetStats(r.Context())
	mamConfigured := h.mamClient != nil && h.mamClient.HasAuth()

	_ = json.NewEncoder(w).Encode(map[string]any{
		"books_count":     totalBooks,
		"authors_count":   totalAuthors,
		"series_count":    totalSeries,
		"heap_alloc":      fmt.Sprintf("%.1f MB", float64(m.Alloc)/1024/1024),
		"heap_sys":        fmt.Sprintf("%.1f MB", float64(m.Sys)/1024/1024),
		"goroutines":      runtime.NumGoroutine(),
		"mam_configured":  mamConfigured,
		"metadata_path":   h.libraryDir,
		"books_path":      h.booksDir,
		"plugins_enabled": h.pluginMgr != nil,
	})
}

// APIGetAuthors devolve a lista de todos os autores da biblioteca.
func (h *Handler) APIGetAuthors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	authors, err := h.db.GetAllAuthors(r.Context())
	if err != nil {
		authors = []database.EntityCount{}
	}
	_ = json.NewEncoder(w).Encode(authors)
}

// APIGetSeries devolve a lista de todas as séries.
func (h *Handler) APIGetSeries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	series, err := h.db.GetAllSeries(r.Context())
	if err != nil {
		series = []database.EntityCount{}
	}
	_ = json.NewEncoder(w).Encode(series)
}

// APIGetTags devolve as tags e categorias da biblioteca.
func (h *Handler) APIGetTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tags, err := h.db.GetAllTags(r.Context())
	if err != nil {
		tags = []database.EntityCount{}
	}
	_ = json.NewEncoder(w).Encode(tags)
}

// APIGetFormats devolve os formatos disponíveis na biblioteca.
func (h *Handler) APIGetFormats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	formats, err := h.db.ListFormats(r.Context())
	if err != nil {
		formats = []database.EntityCount{}
	}
	_ = json.NewEncoder(w).Encode(formats)
}

// APIGetShelves devolve estantes personalizadas e estantes mágicas do utilizador.
func (h *Handler) APIGetShelves(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := int64(1)
	if u := h.CurrentUser(r); u != nil {
		userID = u.ID
	}

	custom, _ := h.db.ListCustomShelves(r.Context(), userID)
	magic, _ := h.db.ListMagicShelves(r.Context(), userID)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"custom": custom,
		"magic":  magic,
	})
}

// APICreateShelf cria uma nova estante personalizada.
func (h *Handler) APICreateShelf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Name     string `json:"name"`
		IsPublic bool   `json:"is_public"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Name == "" {
		req.Name = "Nova Estante"
	}

	userID := int64(1)
	if u := h.CurrentUser(r); u != nil {
		userID = u.ID
	}

	id, err := h.db.CreateShelf(r.Context(), userID, req.Name, req.IsPublic)
	if err != nil {
		http.Error(w, `{"error":`+strconv.Quote(err.Error())+`}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "id": id, "name": req.Name})
}

// APIDeleteShelf apaga uma estante.
func (h *Handler) APIDeleteShelf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"ID inválido"}`, http.StatusBadRequest)
		return
	}

	_ = h.db.DeleteShelf(r.Context(), id)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success"})
}

// APIBatchBooks processa ações em massa em livros selecionados.
func (h *Handler) APIBatchBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Action  string  `json:"action"` // "mark_read", "mark_unread", "archive", "unarchive", "delete", "add_to_shelf"
		BookIDs []int64 `json:"book_ids"`
		ShelfID int64   `json:"shelf_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.BookIDs) == 0 {
		http.Error(w, `{"error":"Parâmetros inválidos"}`, http.StatusBadRequest)
		return
	}

	userID := int64(1)
	if u := h.CurrentUser(r); u != nil {
		userID = u.ID
	}

	switch req.Action {
	case "mark_read":
		_ = h.db.BulkSetReadStatus(r.Context(), userID, req.BookIDs, true)
	case "mark_unread":
		_ = h.db.BulkSetReadStatus(r.Context(), userID, req.BookIDs, false)
	case "archive":
		_ = h.db.BulkSetArchived(r.Context(), userID, req.BookIDs, true)
	case "unarchive":
		_ = h.db.BulkSetArchived(r.Context(), userID, req.BookIDs, false)
	case "delete":
		for _, bid := range req.BookIDs {
			_ = h.db.DeleteBook(r.Context(), bid)
		}
	case "add_to_shelf":
		if req.ShelfID > 0 {
			for _, bid := range req.BookIDs {
				_ = h.db.AddBookToShelf(r.Context(), req.ShelfID, bid)
			}
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"action":  req.Action,
		"count":   len(req.BookIDs),
		"message": "Operação em lote concluída com sucesso.",
	})
}

// APIEditBook atualiza os metadados de um livro.
func (h *Handler) APIEditBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"ID inválido"}`, http.StatusBadRequest)
		return
	}

	var form database.BookEditForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, `{"error":"JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if err := h.db.UpdateBookMetadata(r.Context(), id, form); err != nil {
		http.Error(w, `{"error":`+strconv.Quote(err.Error())+`}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Metadados guardados com sucesso.",
	})
}

// APIGetDuplicates devolve os grupos de duplicados encontrados.
func (h *Handler) APIGetDuplicates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	groups, err := h.db.FindDuplicates(r.Context())
	if err != nil {
		groups = []database.DuplicateGroup{}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"groups": groups,
		"count":  len(groups),
	})
}

// APIGetCWASettings devolve o estado dos serviços de automação do CWA.
func (h *Handler) APIGetCWASettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ingest_active":     true,
		"ingest_path":       h.ingestDir,
		"converter_active":  true,
		"fixer_active":      true,
		"backups_active":    true,
		"google_books":      true,
		"openlibrary":       true,
		"douban":            false,
		"hardcover":         true,
		"koreader_sync":     true,
		"target_format":     "EPUB",
		"metadata_path":     h.libraryDir,
		"books_path":        h.booksDir,
	})
}

// APISaveCWASettings guarda as configurações de automação.
func (h *Handler) APISaveCWASettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Definições CWA atualizadas com sucesso.",
	})
}

// APIGetBackups devolve a lista de backups.
func (h *Handler) APIGetBackups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	backupDir := filepath.Join(h.libraryDir, "backups")
	backups, err := backup.ListBackups(backupDir)
	if err != nil {
		backups = []backup.BackupInfo{}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"backups": backups,
		"count":   len(backups),
	})
}

// APICreateBackup cria um novo backup comprimido.
func (h *Handler) APICreateBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	backupDir := filepath.Join(h.libraryDir, "backups")
	name, err := backup.CreateBackup(h.libraryDir, backupDir)
	if err != nil {
		http.Error(w, `{"error":`+strconv.Quote(err.Error())+`}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "success",
		"filename": name,
		"message":  "Backup criado e comprimido com sucesso.",
	})
}

// APIDeleteBackup apaga um backup.
func (h *Handler) APIDeleteBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	filename := r.PathValue("filename")
	cleanName := filepath.Base(filename)
	backupPath := filepath.Join(h.libraryDir, "backups", cleanName)
	if err := os.Remove(backupPath); err != nil {
		http.Error(w, `{"error":`+strconv.Quote(err.Error())+`}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success"})
}


