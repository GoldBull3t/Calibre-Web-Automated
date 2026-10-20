package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/flosch/pongo2/v6"
)

// ListBooks serve a estante principal e pesquisas com layout idêntico ao Calibre-Web original.
func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		pageStr := r.PathValue("page")
		if pageStr != "" {
			page, _ = strconv.Atoi(pageStr)
		}
	}
	search := strings.TrimSpace(r.URL.Query().Get("query"))
	sortParam := r.URL.Query().Get("sort")
	authorID, _ := strconv.ParseInt(r.URL.Query().Get("author"), 10, 64)
	seriesID, _ := strconv.ParseInt(r.URL.Query().Get("series"), 10, 64)
	tagID, _ := strconv.ParseInt(r.URL.Query().Get("tag"), 10, 64)
	pubID, _ := strconv.ParseInt(r.URL.Query().Get("publisher"), 10, 64)
	ratingID, _ := strconv.ParseInt(r.URL.Query().Get("rating"), 10, 64)
	langID, _ := strconv.ParseInt(r.URL.Query().Get("language"), 10, 64)
	formatParam := r.URL.Query().Get("format")

	readStatus := 0
	title := "Books"
	activeNav := "new"

	path := r.URL.Path
	if strings.HasPrefix(path, "/hot") {
		sortParam = "hotdesc"
		title = "Hot Books"
		activeNav = "hot"
	} else if strings.HasPrefix(path, "/discover") {
		sortParam = "random"
		title = "Discover (Random Books)"
		activeNav = "rand"
	} else if strings.HasPrefix(path, "/newest") {
		sortParam = "new"
		title = "Newest Books"
		activeNav = "new"
	} else if strings.HasPrefix(path, "/readbooks") {
		readStatus = 1
		title = "Read Books"
		activeNav = "read"
	} else if strings.HasPrefix(path, "/unreadbooks") {
		readStatus = 2
		title = "Unread Books"
		activeNav = "unread"
	} else if search != "" {
		title = fmt.Sprintf("Search: %s", search)
		activeNav = ""
	}

	var userID int64
	u := h.CurrentUser(r)
	if u != nil {
		userID = u.ID
	}

	res, err := h.db.ListBooks(r.Context(), database.BookFilter{
		Page:        page,
		PageSize:    24,
		Search:      search,
		Sort:        sortParam,
		AuthorID:    authorID,
		SeriesID:    seriesID,
		TagID:       tagID,
		PublisherID: pubID,
		RatingID:    ratingID,
		LanguageID:  langID,
		Format:      formatParam,
		ReadStatus:  readStatus,
		UserID:      userID,
	})
	if err != nil {
		http.Error(w, "Erro ao consultar biblioteca", http.StatusInternalServerError)
		return
	}

	totalBooks, _, _, _ := h.db.GetStats(r.Context())

	ctx := pongo2.Context{
		"result":      res,
		"query":       search,
		"sort":        sortParam,
		"totalBooks":  totalBooks,
		"title":       title,
		"activeNav":   activeNav,
		"currentUser": u,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.gridTpl.ExecuteWriter(ctx, w)
}

// BrowseUniversal implementa o esquema universal do Calibre-Web: /<data>/<sort_param>/<book_id>/<page>
func (h *Handler) BrowseUniversal(w http.ResponseWriter, r *http.Request) {
	data := strings.ToLower(r.PathValue("data"))
	sortParam := r.PathValue("sort_param")
	idStr := r.PathValue("book_id")
	pageStr := r.PathValue("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}

	// Se não tiver ID específico, redireciona ou chama listagens diretas
	if idStr == "" || idStr == "1" && (data == "author" || data == "series" || data == "category" || data == "publisher" || data == "ratings" || data == "formats" || data == "language") && !strings.Contains(r.URL.Path, "/1") {
		switch data {
		case "author":
			h.ListAuthors(w, r)
			return
		case "series":
			h.ListSeries(w, r)
			return
		case "category":
			h.ListTags(w, r)
			return
		case "publisher":
			h.ListPublishers(w, r)
			return
		case "ratings":
			h.ListRatings(w, r)
			return
		case "formats":
			h.ListFormats(w, r)
			return
		case "language":
			h.ListLanguages(w, r)
			return
		case "hot":
			h.ListBooks(w, r)
			return
		case "discover":
			h.ListBooks(w, r)
			return
		case "newest":
			h.ListBooks(w, r)
			return
		case "readbooks":
			h.ListBooks(w, r)
			return
		case "unreadbooks":
			h.ListBooks(w, r)
			return
		}
	}

	filter := database.BookFilter{
		Page:     page,
		PageSize: 24,
		Sort:     sortParam,
	}

	title := "Books"
	activeNav := ""

	id, _ := strconv.ParseInt(idStr, 10, 64)

	switch data {
	case "author":
		activeNav = "author"
		if id > 0 {
			filter.AuthorID = id
			if author, _ := h.db.GetAuthorByID(r.Context(), id); author != nil {
				title = "Author: " + author.Name
			}
		} else if idStr != "" {
			filter.Search = idStr
			title = "Author: " + idStr
		}
	case "series":
		activeNav = "series"
		if id > 0 {
			filter.SeriesID = id
			if s, _ := h.db.GetSeriesByID(r.Context(), id); s != nil {
				title = "Series: " + s.Name
			}
		} else if idStr != "" {
			filter.Search = idStr
			title = "Series: " + idStr
		}
	case "category":
		activeNav = "category"
		if id > 0 {
			filter.TagID = id
			if t, _ := h.db.GetTagByID(r.Context(), id); t != nil {
				title = "Category: " + t.Name
			}
		} else if idStr != "" {
			filter.Search = idStr
			title = "Category: " + idStr
		}
	case "publisher":
		activeNav = "publisher"
		if id > 0 {
			filter.PublisherID = id
			if p, _ := h.db.GetPublisherByID(r.Context(), id); p != nil {
				title = "Publisher: " + p.Name
			}
		} else if idStr != "" {
			filter.Search = idStr
			title = "Publisher: " + idStr
		}
	case "ratings":
		activeNav = "ratings"
		filter.RatingID = id
		title = fmt.Sprintf("Rating: %d Stars", id)
	case "formats":
		activeNav = "format"
		filter.Format = idStr
		title = "Format: " + strings.ToUpper(idStr)
	case "language":
		activeNav = "language"
		filter.LanguageID = id
		title = "Language: " + idStr
	case "hot":
		activeNav = "hot"
		filter.Sort = "hotdesc"
		title = "Hot Books"
	case "discover":
		activeNav = "rand"
		filter.Sort = "random"
		title = "Discover (Random Books)"
	case "newest":
		activeNav = "new"
		filter.Sort = "new"
		title = "Newest Books"
	case "readbooks":
		activeNav = "read"
		filter.ReadStatus = 1
		title = "Read Books"
	case "unreadbooks":
		activeNav = "unread"
		filter.ReadStatus = 2
		title = "Unread Books"
	case "shelf":
		activeNav = "shelf"
		if id > 0 {
			shelf, _ := h.db.GetShelfByID(r.Context(), id)
			if shelf != nil {
				title = "Shelf: " + shelf.Name
			}
			res, err := h.db.GetShelfBooks(r.Context(), id, sortParam, page, 24)
			if err == nil {
				totalBooks, _, _, _ := h.db.GetStats(r.Context())
				_ = h.gridTpl.ExecuteWriter(pongo2.Context{
					"result":     res,
					"title":      title,
					"activeNav":  activeNav,
					"totalBooks": totalBooks,
					"user":       h.CurrentUser(r),
				}, w)
				return
			}
		}
	case "magicshelf":
		activeNav = "magicshelf"
		if id > 0 {
			userID := int64(1)
			if u := h.CurrentUser(r); u != nil {
				userID = u.ID
			}
			res, err := h.db.GetMagicShelfBooks(r.Context(), id, userID, page, 24)
			if err == nil {
				totalBooks, _, _, _ := h.db.GetStats(r.Context())
				_ = h.gridTpl.ExecuteWriter(pongo2.Context{
					"result":     res,
					"title":      "Magic Shelf",
					"activeNav":  activeNav,
					"totalBooks": totalBooks,
					"user":       h.CurrentUser(r),
				}, w)
				return
			}
		}
	default:
		filter.Search = idStr
		title = "Books"
	}

	u := h.CurrentUser(r)
	if u != nil {
		filter.UserID = u.ID
	}

	res, err := h.db.ListBooks(r.Context(), filter)
	if err != nil {
		http.Error(w, "Erro ao consultar biblioteca", http.StatusInternalServerError)
		return
	}

	totalBooks, _, _, _ := h.db.GetStats(r.Context())

	ctx := pongo2.Context{
		"result":      res,
		"query":       filter.Search,
		"sort":        sortParam,
		"totalBooks":  totalBooks,
		"title":       title,
		"activeNav":   activeNav,
		"currentUser": u,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.gridTpl.ExecuteWriter(ctx, w)
}

// ShowBook serve a página ou fragmento modal com os detalhes de um livro específico.
func (h *Handler) ShowBook(w http.ResponseWriter, r *http.Request) {
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

	u := h.CurrentUser(r)
	if u != nil {
		isRead, isArchived, _ := h.db.GetBookUserState(r.Context(), u.ID, id)
		book.ReadStatus = isRead
		book.IsArchived = isArchived
	}

	isXHR := r.Header.Get("X-Requested-With") == "XMLHttpRequest" || r.URL.Query().Get("modal") == "1"

	ctx := pongo2.Context{
		"book":        book,
		"safeDesc":    template.HTML(book.Comment),
		"currentUser": u,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if isXHR {
		_ = h.modalTpl.ExecuteWriter(ctx, w)
	} else {
		ctx["title"] = book.Title
		_ = h.detailTpl.ExecuteWriter(ctx, w)
	}
}

// ListAuthors lista os autores com paginação.
func (h *Handler) ListAuthors(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	res, err := h.db.ListAuthors(r.Context(), page, 24)
	if err != nil {
		http.Error(w, "Erro ao listar autores", http.StatusInternalServerError)
		return
	}

	type displayItem struct {
		Name  string
		Count int
		URL   string
		Icon  string
	}

	items := make([]displayItem, 0, len(res.Items))
	for _, it := range res.Items {
		items = append(items, displayItem{
			Name:  it.Name,
			Count: it.Count,
			URL:   fmt.Sprintf("/author/stored/%d", it.ID),
			Icon:  "glyphicon-user",
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.entitiesTpl.ExecuteWriter(pongo2.Context{
		"title":      "Authors",
		"activeNav":  "author",
		"items":      items,
		"page":       res.Page,
		"totalPages": res.TotalPages,
		"totalItems": res.TotalItems,
		"basePath":   "/author",
	}, w)
}

// ListSeries lista as séries com paginação.
func (h *Handler) ListSeries(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	res, err := h.db.ListSeries(r.Context(), page, 24)
	if err != nil {
		http.Error(w, "Erro ao listar séries", http.StatusInternalServerError)
		return
	}

	type displayItem struct {
		Name  string
		Count int
		URL   string
		Icon  string
	}

	items := make([]displayItem, 0, len(res.Items))
	for _, it := range res.Items {
		items = append(items, displayItem{
			Name:  it.Name,
			Count: it.Count,
			URL:   fmt.Sprintf("/series/stored/%d", it.ID),
			Icon:  "glyphicon-list",
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.entitiesTpl.ExecuteWriter(pongo2.Context{
		"title":      "Series",
		"activeNav":  "series",
		"items":      items,
		"page":       res.Page,
		"totalPages": res.TotalPages,
		"totalItems": res.TotalItems,
		"basePath":   "/series",
	}, w)
}

// ListTags lista categorias com paginação.
func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	res, err := h.db.ListTags(r.Context(), page, 24)
	if err != nil {
		http.Error(w, "Erro ao listar categorias", http.StatusInternalServerError)
		return
	}

	type displayItem struct {
		Name  string
		Count int
		URL   string
		Icon  string
	}

	items := make([]displayItem, 0, len(res.Items))
	for _, it := range res.Items {
		items = append(items, displayItem{
			Name:  it.Name,
			Count: it.Count,
			URL:   fmt.Sprintf("/category/stored/%d", it.ID),
			Icon:  "glyphicon-tag",
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.entitiesTpl.ExecuteWriter(pongo2.Context{
		"title":      "Categories",
		"activeNav":  "category",
		"items":      items,
		"page":       res.Page,
		"totalPages": res.TotalPages,
		"totalItems": res.TotalItems,
		"basePath":   "/category",
	}, w)
}

// ListPublishers lista editoras com paginação.
func (h *Handler) ListPublishers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	res, err := h.db.ListPublishers(r.Context(), page, 24)
	if err != nil {
		http.Error(w, "Erro ao listar editoras", http.StatusInternalServerError)
		return
	}

	type displayItem struct {
		Name  string
		Count int
		URL   string
		Icon  string
	}

	items := make([]displayItem, 0, len(res.Items))
	for _, it := range res.Items {
		items = append(items, displayItem{
			Name:  it.Name,
			Count: it.Count,
			URL:   fmt.Sprintf("/publisher/stored/%d", it.ID),
			Icon:  "glyphicon-briefcase",
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.entitiesTpl.ExecuteWriter(pongo2.Context{
		"title":      "Publishers",
		"activeNav":  "publisher",
		"items":      items,
		"page":       res.Page,
		"totalPages": res.TotalPages,
		"totalItems": res.TotalItems,
		"basePath":   "/publisher",
	}, w)
}

// ListRatings lista avaliações com paginação.
func (h *Handler) ListRatings(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	res, err := h.db.ListRatings(r.Context(), page, 24)
	if err != nil {
		http.Error(w, "Erro ao listar avaliações", http.StatusInternalServerError)
		return
	}

	type displayItem struct {
		Name  string
		Count int
		URL   string
		Icon  string
	}

	items := make([]displayItem, 0, len(res.Items))
	for _, it := range res.Items {
		items = append(items, displayItem{
			Name:  it.Name,
			Count: it.Count,
			URL:   fmt.Sprintf("/ratings/stored/%d", it.ID),
			Icon:  "glyphicon-star",
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.entitiesTpl.ExecuteWriter(pongo2.Context{
		"title":      "Ratings",
		"activeNav":  "ratings",
		"items":      items,
		"page":       res.Page,
		"totalPages": res.TotalPages,
		"totalItems": res.TotalItems,
		"basePath":   "/ratings",
	}, w)
}

// ListLanguages lista línguas com paginação.
func (h *Handler) ListLanguages(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	res, err := h.db.ListLanguages(r.Context(), page, 24)
	if err != nil {
		http.Error(w, "Erro ao listar idiomas", http.StatusInternalServerError)
		return
	}

	type displayItem struct {
		Name  string
		Count int
		URL   string
		Icon  string
	}

	items := make([]displayItem, 0, len(res.Items))
	for _, it := range res.Items {
		items = append(items, displayItem{
			Name:  it.Name,
			Count: it.Count,
			URL:   fmt.Sprintf("/language/stored/%d", it.ID),
			Icon:  "glyphicon-globe",
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.entitiesTpl.ExecuteWriter(pongo2.Context{
		"title":      "Languages",
		"activeNav":  "language",
		"items":      items,
		"page":       res.Page,
		"totalPages": res.TotalPages,
		"totalItems": res.TotalItems,
		"basePath":   "/language",
	}, w)
}

// ListFormats lista os formatos disponíveis na biblioteca.
func (h *Handler) ListFormats(w http.ResponseWriter, r *http.Request) {
	formats, err := h.db.ListFormats(r.Context())
	if err != nil {
		http.Error(w, "Erro ao listar formatos", http.StatusInternalServerError)
		return
	}

	type displayItem struct {
		Name  string
		Count int
		URL   string
		Icon  string
	}

	items := make([]displayItem, 0, len(formats))
	for _, it := range formats {
		items = append(items, displayItem{
			Name:  it.Name,
			Count: it.Count,
			URL:   fmt.Sprintf("/formats/stored/%s", strings.ToLower(it.Name)),
			Icon:  "glyphicon-file",
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.entitiesTpl.ExecuteWriter(pongo2.Context{
		"title":      "Formats",
		"activeNav":  "format",
		"items":      items,
		"page":       1,
		"totalPages": 1,
		"totalItems": int64(len(items)),
		"basePath":   "/formats",
	}, w)
}

// CreateShelf cria uma nova estante pessoal ou partilhada.
func (h *Handler) CreateShelf(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		title := strings.TrimSpace(r.FormValue("title"))
		if title == "" {
			title = "New Shelf"
		}
		isPublic := r.FormValue("is_public") == "on" || r.FormValue("is_public") == "true"
		id, err := h.db.CreateShelf(r.Context(), 1, title, isPublic)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/shelf/%d", id), http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.shelfEditTpl.ExecuteWriter(pongo2.Context{
		"title": "Create Shelf",
		"shelf": &database.Shelf{Name: "", IsPublic: true},
	}, w)
}

// EditShelf edita uma estante existente.
func (h *Handler) EditShelf(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		title := strings.TrimSpace(r.FormValue("title"))
		if title == "" {
			title = fmt.Sprintf("Shelf %d", id)
		}
		isPublic := r.FormValue("is_public") == "on" || r.FormValue("is_public") == "true"
		_ = h.db.UpdateShelf(r.Context(), id, title, isPublic)
		http.Redirect(w, r, fmt.Sprintf("/shelf/%d", id), http.StatusSeeOther)
		return
	}

	shelf, err := h.db.GetShelfByID(r.Context(), id)
	if err != nil || shelf == nil {
		shelf = &database.Shelf{ID: id, Name: fmt.Sprintf("Shelf %d", id), IsPublic: true}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.shelfEditTpl.ExecuteWriter(pongo2.Context{
		"title": "Edit Shelf",
		"shelf": shelf,
	}, w)
}

// DeleteShelf remove uma estante e seus links.
func (h *Handler) DeleteShelf(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	_ = h.db.DeleteShelf(r.Context(), id)
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// AddToShelf associa um livro à estante.
func (h *Handler) AddToShelf(w http.ResponseWriter, r *http.Request) {
	shelfID, _ := strconv.ParseInt(r.PathValue("shelf_id"), 10, 64)
	bookID, _ := strconv.ParseInt(r.PathValue("book_id"), 10, 64)

	if shelfID > 0 && bookID > 0 {
		_ = h.db.AddBookToShelf(r.Context(), shelfID, bookID)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// RemoveFromShelf remove um livro da estante.
func (h *Handler) RemoveFromShelf(w http.ResponseWriter, r *http.Request) {
	shelfID, _ := strconv.ParseInt(r.PathValue("shelf_id"), 10, 64)
	bookID, _ := strconv.ParseInt(r.PathValue("book_id"), 10, 64)

	if shelfID > 0 && bookID > 0 {
		_ = h.db.RemoveBookFromShelf(r.Context(), shelfID, bookID)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// AddSelectedToShelf adiciona livros selecionados à estante.
func (h *Handler) AddSelectedToShelf(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ShelfID int64   `json:"shelf_id"`
		BookIDs []int64 `json:"book_ids"`
	}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&payload)
	} else {
		_ = r.ParseForm()
		payload.ShelfID, _ = strconv.ParseInt(r.FormValue("shelf_id"), 10, 64)
		bookIDsStr := r.Form["book_ids[]"]
		if len(bookIDsStr) == 0 {
			bookIDsStr = r.Form["book_ids"]
		}
		for _, bStr := range bookIDsStr {
			if bid, err := strconv.ParseInt(bStr, 10, 64); err == nil && bid > 0 {
				payload.BookIDs = append(payload.BookIDs, bid)
			}
		}
	}

	added := 0
	if payload.ShelfID > 0 {
		for _, bid := range payload.BookIDs {
			if err := h.db.AddBookToShelf(r.Context(), payload.ShelfID, bid); err == nil {
				added++
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "added_count": added})
}

// MassAddToShelf adiciona livros em massa a uma estante.
func (h *Handler) MassAddToShelf(w http.ResponseWriter, r *http.Request) {
	shelfID, _ := strconv.ParseInt(r.PathValue("shelf_id"), 10, 64)
	if shelfID <= 0 {
		shelfID, _ = strconv.ParseInt(r.FormValue("shelf_id"), 10, 64)
	}

	var bookIDs []int64
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var payload struct {
			BookIDs []int64 `json:"book_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
			bookIDs = payload.BookIDs
		}
	} else {
		_ = r.ParseForm()
		for _, bStr := range r.Form["book_ids[]"] {
			if bid, err := strconv.ParseInt(bStr, 10, 64); err == nil && bid > 0 {
				bookIDs = append(bookIDs, bid)
			}
		}
	}

	added := 0
	if shelfID > 0 {
		for _, bid := range bookIDs {
			if err := h.db.AddBookToShelf(r.Context(), shelfID, bid); err == nil {
				added++
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "added_count": added})
}

// OrderShelf reordena livros numa estante.
func (h *Handler) OrderShelf(w http.ResponseWriter, r *http.Request) {
	shelfID, _ := strconv.ParseInt(r.PathValue("shelf_id"), 10, 64)

	if r.Method == http.MethodPost {
		orderMap := make(map[int64]int)
		_ = r.ParseForm()
		for key, vals := range r.Form {
			if bookID, err := strconv.ParseInt(key, 10, 64); err == nil && len(vals) > 0 {
				if ord, err := strconv.Atoi(vals[0]); err == nil {
					orderMap[bookID] = ord
				}
			}
		}

		if len(orderMap) > 0 && shelfID > 0 {
			_ = h.db.ReorderShelfBooks(r.Context(), shelfID, orderMap)
		}

		if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success"})
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/shelf/%d", shelfID), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/shelf/%d", shelfID), http.StatusSeeOther)
}


