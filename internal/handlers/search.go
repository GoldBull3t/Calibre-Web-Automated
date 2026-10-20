package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/flosch/pongo2/v6"
)

// Search processa pesquisas simples via query parameter.
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	h.ListBooks(w, r)
}

// AdvSearch renderiza o formulário de pesquisa avançada em GET e executa a pesquisa em POST.
func (h *Handler) AdvSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		authors, _ := h.db.GetAllAuthors(r.Context())
		series, _ := h.db.GetAllSeries(r.Context())
		tags, _ := h.db.GetAllTags(r.Context())
		formats, _ := h.db.ListFormats(r.Context())

		ctx := pongo2.Context{
			"authors": authors,
			"series":  series,
			"tags":    tags,
			"formats": formats,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = h.searchFormTpl.ExecuteWriter(ctx, w)
		return
	}

	// Método POST
	_ = r.ParseForm()
	title := strings.TrimSpace(r.FormValue("title"))
	authorID, _ := strconv.ParseInt(r.FormValue("author"), 10, 64)
	seriesID, _ := strconv.ParseInt(r.FormValue("series"), 10, 64)
	tagID, _ := strconv.ParseInt(r.FormValue("tag"), 10, 64)
	format := strings.TrimSpace(r.FormValue("format"))

	res, err := h.db.ListBooks(r.Context(), database.BookFilter{
		Page:     1,
		PageSize: 24,
		Search:   title,
		AuthorID: authorID,
		SeriesID: seriesID,
		TagID:    tagID,
		Format:   format,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalBooks, _, _, _ := h.db.GetStats(r.Context())

	ctx := pongo2.Context{
		"result":     res,
		"query":      title,
		"totalBooks": totalBooks,
		"title":      "Search Results",
		"activeNav":  "",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.gridTpl.ExecuteWriter(ctx, w)
}

// GetAuthorsJSON devolve os autores em formato JSON para os campos autocomplete typeahead.
func (h *Handler) GetAuthorsJSON(w http.ResponseWriter, r *http.Request) {
	authors, err := h.db.GetAllAuthors(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type jsonEntry struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	res := make([]jsonEntry, 0, len(authors))
	for _, a := range authors {
		res = append(res, jsonEntry{ID: a.ID, Name: a.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// GetPublishersJSON devolve as editoras em formato JSON para os campos autocomplete.
func (h *Handler) GetPublishersJSON(w http.ResponseWriter, r *http.Request) {
	pubs, err := h.db.GetAllPublishers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type jsonEntry struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	res := make([]jsonEntry, 0, len(pubs))
	for _, p := range pubs {
		res = append(res, jsonEntry{ID: p.ID, Name: p.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// GetTagsJSON devolve as tags em formato JSON.
func (h *Handler) GetTagsJSON(w http.ResponseWriter, r *http.Request) {
	tags, err := h.db.GetAllTags(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type jsonEntry struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	res := make([]jsonEntry, 0, len(tags))
	for _, t := range tags {
		res = append(res, jsonEntry{ID: t.ID, Name: t.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// GetSeriesJSON devolve as séries em formato JSON.
func (h *Handler) GetSeriesJSON(w http.ResponseWriter, r *http.Request) {
	series, err := h.db.GetAllSeries(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type jsonEntry struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	res := make([]jsonEntry, 0, len(series))
	for _, s := range series {
		res = append(res, jsonEntry{ID: s.ID, Name: s.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// GetLanguagesJSON devolve as línguas em formato JSON.
func (h *Handler) GetLanguagesJSON(w http.ResponseWriter, r *http.Request) {
	langs, err := h.db.GetAllLanguages(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type jsonEntry struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	res := make([]jsonEntry, 0, len(langs))
	for _, l := range langs {
		res = append(res, jsonEntry{ID: l.ID, Name: l.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// GetMatchingTags responde a pesquisas dinâmicas do seletor de tags.
func (h *Handler) GetMatchingTags(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	tags, err := h.db.GetAllTags(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var matched []string
	for _, t := range tags {
		if q == "" || strings.Contains(strings.ToLower(t.Name), q) {
			matched = append(matched, t.Name)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(matched)
}

