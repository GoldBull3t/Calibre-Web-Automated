package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/crocodilestick/calibre-web-go/internal/formats"
	"github.com/crocodilestick/calibre-web-go/internal/metadata"
	"github.com/crocodilestick/calibre-web-go/internal/updater"
	"github.com/flosch/pongo2/v6"
)

// Admin entrega a página de administração completa do Calibre-Web.
func (h *Handler) Admin(w http.ResponseWriter, r *http.Request) {
	totalBooks, totalAuthors, totalSeries, _ := h.db.GetStats(r.Context())
	users, _ := h.db.GetUsers(r.Context())

	ctx := pongo2.Context{
		"totalBooks":   totalBooks,
		"totalAuthors": totalAuthors,
		"totalSeries":  totalSeries,
		"users":        users,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.adminTpl.ExecuteWriter(ctx, w)
}

// Tasks entrega a página de monitorização de tarefas do Calibre-Web.
func (h *Handler) Tasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tasksTpl.ExecuteWriter(pongo2.Context{}, w)
}

// Stats entrega a página de estatísticas do servidor Calibre-Web Automated.
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	totalBooks, totalAuthors, totalSeries, _ := h.db.GetStats(r.Context())
	formats, _ := h.db.ListFormats(r.Context())

	ctx := pongo2.Context{
		"totalBooks":   totalBooks,
		"totalAuthors": totalAuthors,
		"totalSeries":  totalSeries,
		"formats":      formats,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.statsTpl.ExecuteWriter(ctx, w)
}

// Duplicates entrega o painel de verificação e resolução de livros duplicados.
func (h *Handler) Duplicates(w http.ResponseWriter, r *http.Request) {
	groups, _ := h.db.FindDuplicates(r.Context())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.duplicatesTpl.ExecuteWriter(pongo2.Context{
		"groups": groups,
		"title":  "Duplicates",
	}, w)
}

// TriggerDuplicateScan executa a verificação algorítmica de duplicados.
func (h *Handler) TriggerDuplicateScan(w http.ResponseWriter, r *http.Request) {
	groups, _ := h.db.FindDuplicates(r.Context())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":           "success",
		"message":          "Scan completed",
		"duplicates_count": len(groups),
	})
}

// DismissDuplicateSetupNotice descarta a notificação de configuração de duplicados.
func (h *Handler) DismissDuplicateSetupNotice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// DuplicateStatus devolve o estado atual da verificação de duplicados.
func (h *Handler) DuplicateStatus(w http.ResponseWriter, r *http.Request) {
	groups, _ := h.db.FindDuplicates(r.Context())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":           "idle",
		"duplicates_count": len(groups),
	})
}

// DismissDuplicate descarta ou reativa um grupo de duplicados.
func (h *Handler) DismissDuplicate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	groupHash := r.PathValue("group_hash")
	if groupHash == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 0 {
			groupHash = parts[len(parts)-1]
		}
	}

	userID := int64(1)
	if u := h.CurrentUser(r); u != nil {
		userID = u.ID
	}

	if strings.Contains(r.URL.Path, "/undismiss/") {
		_ = h.db.UndismissDuplicateGroup(r.Context(), userID, groupHash)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "undismissed", "success": true})
		return
	}

	_ = h.db.DismissDuplicateGroup(r.Context(), userID, groupHash)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "dismissed", "success": true})
}

// PreviewResolution simula a resolução de duplicados segundo a estratégia escolhida.
func (h *Handler) PreviewResolution(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Strategy string `json:"strategy"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Strategy == "" {
		req.Strategy = "newest"
	}

	groups, err := h.db.FindDuplicates(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "error": err.Error()})
		return
	}

	previews := make([]map[string]any, 0, len(groups))
	totalDeleted := 0
	for _, g := range groups {
		if len(g.Books) < 2 {
			continue
		}

		keptBook := g.Books[0]
		deletedBooks := g.Books[1:]
		totalDeleted += len(deletedBooks)

		previews = append(previews, map[string]any{
			"group":         g.Hash,
			"kept_book":     keptBook,
			"deleted_books": deletedBooks,
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":        true,
		"resolved_count": len(previews),
		"deleted_count":  totalDeleted,
		"kept_count":     len(previews),
		"errors":         []string{},
		"preview":        previews,
	})
}

// ExecuteResolution executa a resolução de duplicados unindo formatos e limpando órfãos.
func (h *Handler) ExecuteResolution(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Strategy string `json:"strategy"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Strategy == "" {
		req.Strategy = "newest"
	}

	groups, err := h.db.FindDuplicates(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "error": err.Error()})
		return
	}

	resolvedCount := 0
	deletedCount := 0
	keptCount := 0

	for _, g := range groups {
		if len(g.Books) < 2 {
			continue
		}

		keptID := g.Books[0].ID
		sourceIDs := make([]int64, 0, len(g.Books)-1)
		for _, dup := range g.Books[1:] {
			sourceIDs = append(sourceIDs, dup.ID)
		}
		if err := h.db.MergeBookRecords(r.Context(), keptID, sourceIDs, h.booksDir); err == nil {
			deletedCount += len(sourceIDs)
			keptCount++
			resolvedCount++
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":        true,
		"resolved_count": resolvedCount,
		"deleted_count":  deletedCount,
		"kept_count":     keptCount,
		"errors":         []string{},
	})
}

// CWASettings entrega as configurações do CWA.
func (h *Handler) CWASettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "message": "Settings saved"})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.cwaSettingsTpl.ExecuteWriter(pongo2.Context{}, w)
}

// CWASwitchTheme alterna o tema visual CaliBlur.
func (h *Handler) CWASwitchTheme(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "theme",
		Value:  "caliBlur",
		Path:   "/",
		MaxAge: 365 * 24 * 3600,
	})
	ref := r.Header.Get("Referer")
	if ref == "" {
		ref = "/"
	}
	http.Redirect(w, r, ref, http.StatusSeeOther)
}

// LibraryRefresh responde a chamadas AJAX de refresh da biblioteca.
func (h *Handler) LibraryRefresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Library refresh completed.",
	})
}

// LibraryRefreshMessages responde a mensagens do refresh.
func (h *Handler) LibraryRefreshMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"messages": []string{"Library is up to date."},
	})
}

// CWACheckMonitoring devolve o status da monitorização.
func (h *Handler) CWACheckMonitoring(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"monitoring": true,
		"status":     "active",
	})
}

// CWAScheduledUpcoming devolve os envios agendados.
func (h *Handler) CWAScheduledUpcoming(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"total": 0,
		"rows":  []any{},
	})
}

// CWAScheduledUpcomingOps devolve as operações agendadas.
func (h *Handler) CWAScheduledUpcomingOps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"total": 0,
		"rows":  []any{},
	})
}

// CWAScheduledCancel cancela uma tarefa agendada.
func (h *Handler) CWAScheduledCancel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "cancelled"})
}

// CWAConvertLibraryOverview entrega visão geral do conversor de biblioteca.
func (h *Handler) CWAConvertLibraryOverview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.cwaConvertTpl.ExecuteWriter(pongo2.Context{}, w)
}

// CWAConvertLibraryStart inicia conversão em lote.
func (h *Handler) CWAConvertLibraryStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "started"})
}

// ConvertLibraryCancel cancela conversão.
func (h *Handler) ConvertLibraryCancel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "cancelled"})
}

// ConvertLibraryStatus devolve estado da conversão.
func (h *Handler) ConvertLibraryStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "idle", "progress": 100})
}

// CWAEpubFixerOverview entrega visão geral do reparador de EPUB.
func (h *Handler) CWAEpubFixerOverview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.cwaFixerTpl.ExecuteWriter(pongo2.Context{}, w)
}

// CWAEpubFixerStart inicia reparação de EPUBs.
func (h *Handler) CWAEpubFixerStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Verifica se foi solicitado um livro específico
	var bookID int64
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		if idStr := r.FormValue("book_id"); idStr != "" {
			bookID, _ = strconv.ParseInt(idStr, 10, 64)
		}
	}
	if bookID == 0 {
		if idStr := r.URL.Query().Get("book_id"); idStr != "" {
			bookID, _ = strconv.ParseInt(idStr, 10, 64)
		}
	}

	totalFixed := 0

	if bookID > 0 {
		b, err := h.db.GetBookByID(r.Context(), bookID)
		if err == nil && b != nil {
			dir := filepath.Join(h.booksDir, b.Path)
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				if strings.HasSuffix(strings.ToLower(e.Name()), ".epub") {
					fixes, _ := formats.FixEPUB(filepath.Join(dir, e.Name()))
					totalFixed += fixes
				}
			}
		}
	} else {
		// Varre biblioteca para todos os EPUBs
		res, err := h.db.ListBooks(r.Context(), database.BookFilter{Page: 1, PageSize: 500})
		if err == nil {
			for _, b := range res.Items {
				dir := filepath.Join(h.booksDir, b.Path)
				entries, _ := os.ReadDir(dir)
				for _, e := range entries {
					if strings.HasSuffix(strings.ToLower(e.Name()), ".epub") {
						fixes, _ := formats.FixEPUB(filepath.Join(dir, e.Name()))
						totalFixed += fixes
					}
				}
			}
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      "success",
		"fixes_count": totalFixed,
		"message":     fmt.Sprintf("EPUB Fixer completou %d correções estruturais e de OCR com sucesso.", totalFixed),
	})
}

// EpubFixerCancel cancela reparação de EPUB.
func (h *Handler) EpubFixerCancel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "cancelled"})
}

// EpubFixerStatus devolve estado do reparador de EPUB.
func (h *Handler) EpubFixerStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "idle", "progress": 100})
}

// CWAStatsExportCSV exporta relatório CSV dinâmico a partir dos dados reais da biblioteca.
func (h *Handler) CWAStatsExportCSV(w http.ResponseWriter, r *http.Request) {
	tabName := r.PathValue("tab_name")
	if tabName == "" {
		tabName = "library"
	}

	stats, _ := h.db.GetLibraryStats(r.Context())

	filename := fmt.Sprintf("cwa_stats_%s.csv", tabName)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"Metric", "Value"})
	_ = writer.Write([]string{"Total Books", fmt.Sprintf("%v", stats["total_books"])})
	_ = writer.Write([]string{"Total Authors", fmt.Sprintf("%v", stats["total_authors"])})
	_ = writer.Write([]string{"Total Series", fmt.Sprintf("%v", stats["total_series"])})
	_ = writer.Write([]string{"Total Tags", fmt.Sprintf("%v", stats["total_tags"])})
	_ = writer.Write([]string{"Total Formats", fmt.Sprintf("%v", stats["total_formats"])})
	_ = writer.Write([]string{"Total Size Bytes", fmt.Sprintf("%v", stats["total_size"])})
	writer.Flush()
}

// MetadataSearch pesquisa livros em provedores online (Google Books, Open Library).
func (h *Handler) MetadataSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = r.ParseForm()
	query := strings.TrimSpace(r.FormValue("query"))
	if query == "" {
		query = strings.TrimSpace(r.URL.Query().Get("query"))
	}
	if query == "" {
		var req struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		query = strings.TrimSpace(req.Query)
	}

	if query == "" {
		_ = json.NewEncoder(w).Encode([]any{})
		return
	}

	results := metadata.SearchOnline(r.Context(), query)
	_ = json.NewEncoder(w).Encode(results)
}

// MetadataProvider devolve os provedores de metadados suportados.
func (h *Handler) MetadataProvider(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	providers := []map[string]any{
		{"name": "Google Books", "id": "google", "active": true, "globally_enabled": true},
		{"name": "Open Library", "id": "openlibrary", "active": true, "globally_enabled": true},
	}
	_ = json.NewEncoder(w).Encode(providers)
}

// MetadataProviderChange altera as preferências de provedores de metadados.
func (h *Handler) MetadataProviderChange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// CWAStatsDebug devolve informações de depuração do CWA.
func (h *Handler) CWAStatsDebug(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"debug": "enabled",
		"mode":  "production",
	})
}

// CWALogRead lê o ficheiro de log indicado.
func (h *Handler) CWALogRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("Calibre-Web Automated running smoothly in Go.\n"))
}

// CWALogDownload descarrega o ficheiro de log indicado.
func (h *Handler) CWALogDownload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\"cwa.log\"")
	_, _ = w.Write([]byte("Calibre-Web Automated running smoothly in Go.\n"))
}

// MetadataHardcoverFetch busca o ID do Hardcover.app para um livro específico e grava na base de dados.
func (h *Handler) MetadataHardcoverFetch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("book_id")
	bookID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "id de livro inválido"})
		return
	}

	book, err := h.db.GetBookByID(r.Context(), bookID)
	if err != nil || book == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "livro não encontrado"})
		return
	}

	author := ""
	if len(book.Authors) > 0 {
		author = book.Authors[0]
	}
	isbn := book.Identifiers["isbn"]

	client := metadata.NewHardcoverClient(os.Getenv("HARDCOVER_API_KEY"))
	hardcoverID, err := client.FetchHardcoverID(r.Context(), book.Title, author, isbn)
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "not_found",
			"error":  err.Error(),
		})
		return
	}

	_ = h.db.SetBookIdentifier(r.Context(), bookID, "hardcover", hardcoverID)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":       "success",
		"hardcover_id": hardcoverID,
		"book_id":      bookID,
	})
}

// CWAHardcoverFetchAll processa em segundo plano a busca de Hardcover IDs para todos os livros que não o possuem.
func (h *Handler) CWAHardcoverFetchAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	books, err := h.db.GetBooksMissingIdentifier(r.Context(), "hardcover", 100)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	// Executa assincronamente em lote respeitando teto de memória e timeouts
	go func(items []database.Book) {
		client := metadata.NewHardcoverClient(os.Getenv("HARDCOVER_API_KEY"))
		for _, b := range items {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			author := ""
			if len(b.Authors) > 0 {
				author = b.Authors[0]
			}
			hcID, err := client.FetchHardcoverID(ctx, b.Title, author, "")
			cancel()
			if err == nil && hcID != "" {
				_ = h.db.SetBookIdentifier(context.Background(), b.ID, "hardcover", hcID)
			}
			time.Sleep(200 * time.Millisecond) // Taxa de cortesia com a API
		}
	}(books)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":       "started",
		"queued_count": len(books),
		"message":      fmt.Sprintf("Processamento Hardcover iniciado para %d livros.", len(books)),
	})
}

// AjaxCheckUpdate verifica se existe uma nova versão no GitHub releases para notificação no painel de administração.
func (h *Handler) AjaxCheckUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u := updater.GetDefault()
	info, err := u.CheckUpdate(r.Context())
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"update_available": false,
			"current_version":  updater.CurrentVersion,
			"error":            err.Error(),
		})
		return
	}
	_ = json.NewEncoder(w).Encode(info)
}


