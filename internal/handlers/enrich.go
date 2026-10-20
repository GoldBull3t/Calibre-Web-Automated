package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/crocodilestick/calibre-web-go/internal/events"
	"github.com/crocodilestick/calibre-web-go/internal/metadata"
)

// APIEnrichSearch procura candidatos online (OpenLibrary, Google Books) para um livro.
func (h *Handler) APIEnrichSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"ID de livro inválido"}`, http.StatusBadRequest)
		return
	}

	book, err := h.db.GetBookByID(r.Context(), id)
	if err != nil || book == nil {
		http.Error(w, `{"error":"Livro não encontrado"}`, http.StatusNotFound)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("query"))
	author := ""
	if len(book.Authors) > 0 {
		author = book.Authors[0]
	}

	candidates, err := metadata.FetchEnrichmentCandidates(r.Context(), query, book.Title, author)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Falha na busca online: %v"}`, err), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":     "ok",
		"book_id":    id,
		"candidates": candidates,
	})
}

// EnrichApplyPayload define o corpo JSON recebido para aplicar metadados e capa.
type EnrichApplyPayload struct {
	CoverURL    string            `json:"cover_url"`
	Title       string            `json:"title"`
	Authors     []string          `json:"authors"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Publisher   string            `json:"publisher"`
	PublishDate string            `json:"publish_date"`
	Rating      int               `json:"rating"`
	Identifiers map[string]string `json:"identifiers"`
}

// APIEnrichApply aplica a capa HD selecionada e atualiza os metadados escolhidos.
func (h *Handler) APIEnrichApply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"ID de livro inválido"}`, http.StatusBadRequest)
		return
	}

	book, err := h.db.GetBookByID(r.Context(), id)
	if err != nil || book == nil {
		http.Error(w, `{"error":"Livro não encontrado"}`, http.StatusNotFound)
		return
	}

	var payload EnrichApplyPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"JSON inválido"}`, http.StatusBadRequest)
		return
	}

	// 1. Descarrega e guarda a capa HD se fornecida
	if strings.HasPrefix(payload.CoverURL, "http://") || strings.HasPrefix(payload.CoverURL, "https://") {
		if err := h.downloadAndSaveCover(r.Context(), book.ID, book.Path, payload.CoverURL); err != nil {
			// Log apenas, não interrompe a gravação de metadados
			fmt.Printf("[ENRICH] Erro ao descarregar capa %s: %v\n", payload.CoverURL, err)
		}
	}

	// 2. Prepara atualização de metadados
	var publishers []string
	if payload.Publisher != "" {
		publishers = []string{payload.Publisher}
	}

	var authors []string
	if len(payload.Authors) > 0 {
		authors = payload.Authors
	} else {
		authors = book.Authors
	}

	title := payload.Title
	if title == "" {
		title = book.Title
	}

	comment := payload.Description
	if comment == "" {
		comment = book.Comment
	}

	tags := payload.Tags
	if len(tags) == 0 {
		tags = book.Tags
	}

	identifiers := make(map[string]string)
	for k, v := range book.Identifiers {
		identifiers[k] = v
	}
	for k, v := range payload.Identifiers {
		identifiers[k] = v
	}

	form := database.BookEditForm{
		Title:       title,
		TitleSort:   title,
		Authors:     authors,
		Series:      book.Series,
		SeriesIndex: book.SeriesIndex,
		Tags:        tags,
		Publishers:  publishers,
		Rating:      payload.Rating,
		Comment:     comment,
		Identifiers: identifiers,
	}

	if err := h.db.UpdateBookMetadata(r.Context(), id, form); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Falha ao atualizar metadados: %v"}`, err), http.StatusInternalServerError)
		return
	}

	events.GetBroker().Broadcast("book.enriched", fmt.Sprintf("'%s' enriquecido com sucesso", title), map[string]any{
		"book_id": id,
		"title":   title,
	})

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"message": "Metadados e capa enriquecidos com sucesso!",
	})
}

// APIEnrichAuto executa o enriquecimento automático inteligente em 1 clique.
func (h *Handler) APIEnrichAuto(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"ID de livro inválido"}`, http.StatusBadRequest)
		return
	}

	book, err := h.db.GetBookByID(r.Context(), id)
	if err != nil || book == nil {
		http.Error(w, `{"error":"Livro não encontrado"}`, http.StatusNotFound)
		return
	}

	author := ""
	if len(book.Authors) > 0 {
		author = book.Authors[0]
	}

	candidates, err := metadata.FetchEnrichmentCandidates(r.Context(), "", book.Title, author)
	if err != nil || len(candidates) == 0 {
		http.Error(w, `{"error":"Nenhum metadado compatível encontrado online"}`, http.StatusNotFound)
		return
	}

	// Seleciona o melhor candidato (prioriza o que tiver capas e sinopse)
	selected := candidates[0]
	for _, c := range candidates {
		if len(c.Covers) > 0 && len(c.Description) > 50 {
			selected = c
			break
		}
	}

	// Aplica a melhor capa se houver
	if len(selected.Covers) > 0 {
		_ = h.downloadAndSaveCover(r.Context(), book.ID, book.Path, selected.Covers[0].URL)
	}

	// Atualiza metadados
	var pubs []string
	if selected.Publisher != "" {
		pubs = []string{selected.Publisher}
	}

	var authors []string
	if len(selected.Authors) > 0 {
		authors = selected.Authors
	} else {
		authors = book.Authors
	}

	title := selected.Title
	if title == "" {
		title = book.Title
	}

	comment := selected.Description
	if comment == "" {
		comment = book.Comment
	}

	tags := selected.Tags
	if len(tags) == 0 {
		tags = book.Tags
	}

	identifiers := make(map[string]string)
	for k, v := range book.Identifiers {
		identifiers[k] = v
	}
	for k, v := range selected.Identifiers {
		identifiers[k] = v
	}

	form := database.BookEditForm{
		Title:       title,
		TitleSort:   title,
		Authors:     authors,
		Series:      book.Series,
		SeriesIndex: book.SeriesIndex,
		Tags:        tags,
		Publishers:  pubs,
		Rating:      int(selected.Rating * 2), // Converte de 5 estrelas para base 10
		Comment:     comment,
		Identifiers: identifiers,
	}

	if err := h.db.UpdateBookMetadata(r.Context(), id, form); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Falha ao atualizar base de dados: %v"}`, err), http.StatusInternalServerError)
		return
	}

	events.GetBroker().Broadcast("book.enriched", fmt.Sprintf("'%s' auto-enriquecido", title), map[string]any{
		"book_id": id,
		"title":   title,
	})

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "ok",
		"message":   "Livro auto-enriquecido com sucesso em 1 clique!",
		"candidate": selected,
	})
}

// APICleanLibraryMetadata sanitiza autores, títulos e comentários de toda a biblioteca.
func (h *Handler) APICleanLibraryMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	res, err := h.db.ListBooks(r.Context(), database.BookFilter{Page: 1, PageSize: 10000})
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Falha ao listar livros: %v"}`, err), http.StatusInternalServerError)
		return
	}
	books := res.Items

	cleanedCount := 0
	for _, b := range books {
		bDetails, err := h.db.GetBookByID(r.Context(), b.ID)
		if err != nil || bDetails == nil {
			continue
		}

		needsUpdate := false

		// 1. Limpa autores
		var newAuthors []string
		for _, a := range bDetails.Authors {
			c := metadata.CleanAuthorName(a)
			if c != a {
				needsUpdate = true
			}
			newAuthors = append(newAuthors, c)
		}

		// 2. Limpa título
		newTitle := metadata.CleanBookTitle(bDetails.Title)
		if newTitle != bDetails.Title {
			needsUpdate = true
		}

		// 3. Limpa sinopse
		newComment := metadata.CleanGutenbergText(bDetails.Comment)
		if newComment != bDetails.Comment {
			needsUpdate = true
		}

		if needsUpdate {
			form := database.BookEditForm{
				Title:       newTitle,
				TitleSort:   newTitle,
				Authors:     newAuthors,
				Series:      bDetails.Series,
				SeriesIndex: bDetails.SeriesIndex,
				Tags:        bDetails.Tags,
				Publishers:  bDetails.Publishers,
				Rating:      bDetails.Rating,
				Comment:     newComment,
				Identifiers: bDetails.Identifiers,
			}
			if err := h.db.UpdateBookMetadata(r.Context(), bDetails.ID, form); err == nil {
				cleanedCount++
			}
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":        "ok",
		"cleaned_count": cleanedCount,
		"total_books":   len(books),
		"message":       fmt.Sprintf("%d obras sanitizadas com sucesso.", cleanedCount),
	})
}

// downloadAndSaveCover faz o download da imagem e guarda como cover.jpg no livro.
func (h *Handler) downloadAndSaveCover(ctx context.Context, bookID int64, bookRelPath, coverURL string) error {
	client := &http.Client{Timeout: 12 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, coverURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Calibre-Web-Automated-Go/2.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status HTTP %d", resp.StatusCode)
	}

	imgData, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil || len(imgData) < 100 {
		return fmt.Errorf("imagem corrompida ou vazia")
	}

	bookDir := filepath.Join(h.booksDir, filepath.FromSlash(bookRelPath))
	_ = os.MkdirAll(bookDir, 0755)
	coverPath := filepath.Join(bookDir, "cover.jpg")

	if err := os.WriteFile(coverPath, imgData, 0644); err != nil {
		return err
	}

	_ = h.db.SetBookHasCover(ctx, bookID, true)

	cacheKey := fmt.Sprintf("cover:%d", bookID)
	if h.coverCache != nil {
		h.coverCache.Delete(cacheKey)
		if len(imgData) <= 512*1024 {
			h.coverCache.Put(cacheKey, imgData, "image/jpeg")
		}
	}

	return nil
}
