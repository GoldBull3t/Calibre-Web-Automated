package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/database"
)

type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Xmlns   string      `xml:"xmlns,attr"`
	ID      string      `xml:"id"`
	Title   string      `xml:"title"`
	Updated string      `xml:"updated"`
	Author  AtomAuthor  `xml:"author"`
	Links   []AtomLink  `xml:"link"`
	Entries []AtomEntry `xml:"entry"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
	URI  string `xml:"uri,omitempty"`
}

type AtomLink struct {
	Rel   string `xml:"rel,attr,omitempty"`
	Href  string `xml:"href,attr"`
	Type  string `xml:"type,attr,omitempty"`
	Title string `xml:"title,attr,omitempty"`
}

type AtomEntry struct {
	Title   string       `xml:"title"`
	ID      string       `xml:"id"`
	Updated string       `xml:"updated"`
	Summary string       `xml:"summary,omitempty"`
	Authors []AtomAuthor `xml:"author,omitempty"`
	Links   []AtomLink   `xml:"link"`
}

// OPDSFeed gera o catálogo Atom/XML ou despacha para subfeeds.
func (h *Handler) OPDSFeed(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/opds")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		h.opdsRootCatalog(w, r)
		return
	}

	parts := strings.Split(path, "/")
	section := parts[0]

	switch section {
	case "osd", "osd.xml":
		h.opdsOpenSearch(w, r)
		return
	case "search":
		query := r.URL.Query().Get("query")
		if len(parts) > 1 && query == "" {
			query = parts[1]
		}
		h.opdsBookFeed(w, r, "Search Results", database.BookFilter{Search: query, PageSize: 50})
		return
	case "new":
		h.opdsBookFeed(w, r, "Newest Books", database.BookFilter{Sort: "new", PageSize: 50})
		return
	case "hot":
		h.opdsBookFeed(w, r, "Hot Books", database.BookFilter{Sort: "hotdesc", PageSize: 50})
		return
	case "discover":
		h.opdsBookFeed(w, r, "Discover (Random)", database.BookFilter{Sort: "random", PageSize: 50})
		return
	case "ratings":
		if len(parts) > 1 {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			h.opdsBookFeed(w, r, fmt.Sprintf("Rating %d", id), database.BookFilter{RatingID: id, PageSize: 50})
			return
		}
		h.opdsRatingsIndex(w, r)
		return
	case "formats":
		if len(parts) > 1 {
			h.opdsBookFeed(w, r, fmt.Sprintf("Format %s", parts[1]), database.BookFilter{Format: parts[1], PageSize: 50})
			return
		}
		h.opdsFormatsIndex(w, r)
		return
	case "author":
		if len(parts) > 1 {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			h.opdsBookFeed(w, r, "Books by Author", database.BookFilter{AuthorID: id, PageSize: 50})
			return
		}
		h.opdsAuthorsIndex(w, r)
		return
	case "series":
		if len(parts) > 1 {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			h.opdsBookFeed(w, r, "Books in Series", database.BookFilter{SeriesID: id, PageSize: 50})
			return
		}
		h.opdsSeriesIndex(w, r)
		return
	case "category":
		if len(parts) > 1 {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			h.opdsBookFeed(w, r, "Books in Category", database.BookFilter{TagID: id, PageSize: 50})
			return
		}
		h.opdsCategoriesIndex(w, r)
		return
	case "publisher":
		if len(parts) > 1 {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			h.opdsBookFeed(w, r, "Books by Publisher", database.BookFilter{PublisherID: id, PageSize: 50})
			return
		}
		h.opdsPublishersIndex(w, r)
		return
	case "language":
		if len(parts) > 1 {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			h.opdsBookFeed(w, r, "Books by Language", database.BookFilter{LanguageID: id, PageSize: 50})
			return
		}
		h.opdsLanguagesIndex(w, r)
		return
	case "readbooks":
		h.opdsBookFeed(w, r, "Read Books", database.BookFilter{ReadStatus: 1, PageSize: 50})
		return
	case "unreadbooks":
		h.opdsBookFeed(w, r, "Unread Books", database.BookFilter{ReadStatus: 2, PageSize: 50})
		return
	case "cover", "cover_240_240", "cover_90_90", "thumb_240_240":
		h.GetCover(w, r)
		return
	case "download":
		h.DownloadBook(w, r)
		return
	default:
		h.opdsBookFeed(w, r, "Books", database.BookFilter{PageSize: 50})
		return
	}
}

func (h *Handler) opdsRootCatalog(w http.ResponseWriter, r *http.Request) {
	feed := AtomFeed{
		Xmlns:   "http://www.w3.org/2005/Atom",
		ID:      "urn:calibre-web:root",
		Title:   "Calibre-Web Catalog",
		Updated: time.Now().UTC().Format(time.RFC3339),
		Author: AtomAuthor{
			Name: "Calibre-Web",
			URI:  "/opds",
		},
		Links: []AtomLink{
			{Rel: "self", Href: "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "search", Href: "/opds/osd.xml", Type: "application/opensearchdescription+xml"},
		},
		Entries: []AtomEntry{
			{
				Title:   "Newest Books",
				ID:      "urn:calibre-web:new",
				Updated: time.Now().UTC().Format(time.RFC3339),
				Summary: "The most recently added books in the library",
				Links: []AtomLink{
					{Rel: "subsection", Href: "/opds/new", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
				},
			},
			{
				Title:   "Hot Books",
				ID:      "urn:calibre-web:hot",
				Updated: time.Now().UTC().Format(time.RFC3339),
				Summary: "Popular and trending books in the library",
				Links: []AtomLink{
					{Rel: "subsection", Href: "/opds/hot", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
				},
			},
			{
				Title:   "Discover",
				ID:      "urn:calibre-web:discover",
				Updated: time.Now().UTC().Format(time.RFC3339),
				Summary: "Random books to discover",
				Links: []AtomLink{
					{Rel: "subsection", Href: "/opds/discover", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
				},
			},
			{
				Title:   "Authors",
				ID:      "urn:calibre-web:authors",
				Updated: time.Now().UTC().Format(time.RFC3339),
				Summary: "Browse books organized by author",
				Links: []AtomLink{
					{Rel: "subsection", Href: "/opds/author", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
				},
			},
			{
				Title:   "Series",
				ID:      "urn:calibre-web:series",
				Updated: time.Now().UTC().Format(time.RFC3339),
				Summary: "Browse books organized by series",
				Links: []AtomLink{
					{Rel: "subsection", Href: "/opds/series", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
				},
			},
			{
				Title:   "Categories",
				ID:      "urn:calibre-web:categories",
				Updated: time.Now().UTC().Format(time.RFC3339),
				Summary: "Browse books organized by tag or category",
				Links: []AtomLink{
					{Rel: "subsection", Href: "/opds/category", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
				},
			},
			{
				Title:   "Formats",
				ID:      "urn:calibre-web:formats",
				Updated: time.Now().UTC().Format(time.RFC3339),
				Summary: "Browse books by available file format",
				Links: []AtomLink{
					{Rel: "subsection", Href: "/opds/formats", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
				},
			},
			{
				Title:   "Top Rated",
				ID:      "urn:calibre-web:ratings",
				Updated: time.Now().UTC().Format(time.RFC3339),
				Summary: "Browse books by star ratings",
				Links: []AtomLink{
					{Rel: "subsection", Href: "/opds/ratings", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
				},
			},
			{
				Title:   "Publishers",
				ID:      "urn:calibre-web:publishers",
				Updated: time.Now().UTC().Format(time.RFC3339),
				Summary: "Browse books by publisher",
				Links: []AtomLink{
					{Rel: "subsection", Href: "/opds/publisher", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
				},
			},
			{
				Title:   "Languages",
				ID:      "urn:calibre-web:languages",
				Updated: time.Now().UTC().Format(time.RFC3339),
				Summary: "Browse books by language",
				Links: []AtomLink{
					{Rel: "subsection", Href: "/opds/language", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
				},
			},
		},
	}

	writeOPDS(w, feed)
}

func (h *Handler) opdsAuthorsIndex(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	res, err := h.db.ListAuthors(r.Context(), page, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed := AtomFeed{
		Xmlns:   "http://www.w3.org/2005/Atom",
		ID:      "urn:calibre-web:authors",
		Title:   "Authors",
		Updated: time.Now().UTC().Format(time.RFC3339),
		Links: []AtomLink{
			{Rel: "self", Href: fmt.Sprintf("/opds/author?page=%d", res.Page), Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}

	for _, a := range res.Items {
		feed.Entries = append(feed.Entries, AtomEntry{
			Title:   fmt.Sprintf("%s (%d)", a.Name, a.Count),
			ID:      fmt.Sprintf("urn:calibre-web:author:%d", a.ID),
			Updated: time.Now().UTC().Format(time.RFC3339),
			Links: []AtomLink{
				{Rel: "subsection", Href: fmt.Sprintf("/opds/author/%d", a.ID), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			},
		})
	}

	writeOPDS(w, feed)
}

func (h *Handler) opdsSeriesIndex(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	res, err := h.db.ListSeries(r.Context(), page, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed := AtomFeed{
		Xmlns:   "http://www.w3.org/2005/Atom",
		ID:      "urn:calibre-web:series",
		Title:   "Series",
		Updated: time.Now().UTC().Format(time.RFC3339),
		Links: []AtomLink{
			{Rel: "self", Href: fmt.Sprintf("/opds/series?page=%d", res.Page), Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}

	for _, s := range res.Items {
		feed.Entries = append(feed.Entries, AtomEntry{
			Title:   fmt.Sprintf("%s (%d)", s.Name, s.Count),
			ID:      fmt.Sprintf("urn:calibre-web:series:%d", s.ID),
			Updated: time.Now().UTC().Format(time.RFC3339),
			Links: []AtomLink{
				{Rel: "subsection", Href: fmt.Sprintf("/opds/series/%d", s.ID), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			},
		})
	}

	writeOPDS(w, feed)
}

func (h *Handler) opdsCategoriesIndex(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	res, err := h.db.ListTags(r.Context(), page, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed := AtomFeed{
		Xmlns:   "http://www.w3.org/2005/Atom",
		ID:      "urn:calibre-web:categories",
		Title:   "Categories",
		Updated: time.Now().UTC().Format(time.RFC3339),
		Links: []AtomLink{
			{Rel: "self", Href: fmt.Sprintf("/opds/category?page=%d", res.Page), Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}

	for _, t := range res.Items {
		feed.Entries = append(feed.Entries, AtomEntry{
			Title:   fmt.Sprintf("%s (%d)", t.Name, t.Count),
			ID:      fmt.Sprintf("urn:calibre-web:tag:%d", t.ID),
			Updated: time.Now().UTC().Format(time.RFC3339),
			Links: []AtomLink{
				{Rel: "subsection", Href: fmt.Sprintf("/opds/category/%d", t.ID), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			},
		})
	}

	writeOPDS(w, feed)
}

func (h *Handler) opdsFormatsIndex(w http.ResponseWriter, r *http.Request) {
	formats, err := h.db.ListFormats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed := AtomFeed{
		Xmlns:   "http://www.w3.org/2005/Atom",
		ID:      "urn:calibre-web:formats",
		Title:   "Formats",
		Updated: time.Now().UTC().Format(time.RFC3339),
		Links: []AtomLink{
			{Rel: "self", Href: "/opds/formats", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}

	for _, f := range formats {
		feed.Entries = append(feed.Entries, AtomEntry{
			Title:   fmt.Sprintf("%s (%d)", f.Name, f.Count),
			ID:      fmt.Sprintf("urn:calibre-web:format:%s", f.Name),
			Updated: time.Now().UTC().Format(time.RFC3339),
			Links: []AtomLink{
				{Rel: "subsection", Href: fmt.Sprintf("/opds/formats/%s", strings.ToLower(f.Name)), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			},
		})
	}

	writeOPDS(w, feed)
}

func (h *Handler) opdsRatingsIndex(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	res, err := h.db.ListRatings(r.Context(), page, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed := AtomFeed{
		Xmlns:   "http://www.w3.org/2005/Atom",
		ID:      "urn:calibre-web:ratings",
		Title:   "Ratings",
		Updated: time.Now().UTC().Format(time.RFC3339),
		Links: []AtomLink{
			{Rel: "self", Href: "/opds/ratings", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}

	for _, r := range res.Items {
		feed.Entries = append(feed.Entries, AtomEntry{
			Title:   fmt.Sprintf("%s (%d)", r.Name, r.Count),
			ID:      fmt.Sprintf("urn:calibre-web:rating:%d", r.ID),
			Updated: time.Now().UTC().Format(time.RFC3339),
			Links: []AtomLink{
				{Rel: "subsection", Href: fmt.Sprintf("/opds/ratings/%d", r.ID), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			},
		})
	}

	writeOPDS(w, feed)
}

func (h *Handler) opdsPublishersIndex(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	res, err := h.db.ListPublishers(r.Context(), page, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed := AtomFeed{
		Xmlns:   "http://www.w3.org/2005/Atom",
		ID:      "urn:calibre-web:publishers",
		Title:   "Publishers",
		Updated: time.Now().UTC().Format(time.RFC3339),
		Links: []AtomLink{
			{Rel: "self", Href: "/opds/publisher", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}

	for _, p := range res.Items {
		feed.Entries = append(feed.Entries, AtomEntry{
			Title:   fmt.Sprintf("%s (%d)", p.Name, p.Count),
			ID:      fmt.Sprintf("urn:calibre-web:publisher:%d", p.ID),
			Updated: time.Now().UTC().Format(time.RFC3339),
			Links: []AtomLink{
				{Rel: "subsection", Href: fmt.Sprintf("/opds/publisher/%d", p.ID), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			},
		})
	}

	writeOPDS(w, feed)
}

func (h *Handler) opdsLanguagesIndex(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	res, err := h.db.ListLanguages(r.Context(), page, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed := AtomFeed{
		Xmlns:   "http://www.w3.org/2005/Atom",
		ID:      "urn:calibre-web:languages",
		Title:   "Languages",
		Updated: time.Now().UTC().Format(time.RFC3339),
		Links: []AtomLink{
			{Rel: "self", Href: "/opds/language", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}

	for _, l := range res.Items {
		feed.Entries = append(feed.Entries, AtomEntry{
			Title:   fmt.Sprintf("%s (%d)", l.Name, l.Count),
			ID:      fmt.Sprintf("urn:calibre-web:language:%d", l.ID),
			Updated: time.Now().UTC().Format(time.RFC3339),
			Links: []AtomLink{
				{Rel: "subsection", Href: fmt.Sprintf("/opds/language/%d", l.ID), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			},
		})
	}

	writeOPDS(w, feed)
}

func (h *Handler) opdsBookFeed(w http.ResponseWriter, r *http.Request, title string, filter database.BookFilter) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	filter.Page = page

	result, err := h.db.ListBooks(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed := AtomFeed{
		Xmlns:   "http://www.w3.org/2005/Atom",
		ID:      fmt.Sprintf("urn:calibre-web:%s", strings.ToLower(strings.ReplaceAll(title, " ", "-"))),
		Title:   title,
		Updated: time.Now().UTC().Format(time.RFC3339),
		Author: AtomAuthor{
			Name: "Calibre-Web",
			URI:  "/opds",
		},
		Links: []AtomLink{
			{Rel: "self", Href: fmt.Sprintf("%s?page=%d", r.URL.Path, page), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			{Rel: "start", Href: "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}

	if page < result.TotalPages {
		feed.Links = append(feed.Links, AtomLink{
			Rel:  "next",
			Href: fmt.Sprintf("%s?page=%d", r.URL.Path, page+1),
			Type: "application/atom+xml;profile=opds-catalog;kind=acquisition",
		})
	}

	for _, b := range result.Items {
		entry := AtomEntry{
			Title:   b.Title,
			ID:      fmt.Sprintf("urn:calibre-book:%d", b.ID),
			Updated: b.PubDate.Format(time.RFC3339),
			Links: []AtomLink{
				{Rel: "http://opds-spec.org/image", Href: fmt.Sprintf("/cover/%d", b.ID), Type: "image/jpeg"},
				{Rel: "http://opds-spec.org/image/thumbnail", Href: fmt.Sprintf("/cover/%d", b.ID), Type: "image/jpeg"},
				{Rel: "alternate", Href: fmt.Sprintf("/book/%d", b.ID), Type: "text/html"},
			},
		}

		for _, a := range b.Authors {
			entry.Authors = append(entry.Authors, AtomAuthor{Name: a})
		}

		for _, f := range b.Formats {
			mime := "application/octet-stream"
			switch strings.ToUpper(f.Format) {
			case "EPUB":
				mime = "application/epub+zip"
			case "PDF":
				mime = "application/pdf"
			case "MOBI":
				mime = "application/x-mobipocket-ebook"
			}
			entry.Links = append(entry.Links, AtomLink{
				Rel:   "http://opds-spec.org/acquisition",
				Href:  fmt.Sprintf("/opds/download/%d/%s/", b.ID, strings.ToLower(f.Format)),
				Type:  mime,
				Title: f.Format,
			})
		}

		feed.Entries = append(feed.Entries, entry)
	}

	writeOPDS(w, feed)
}

func (h *Handler) opdsOpenSearch(w http.ResponseWriter, r *http.Request) {
	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/">
   <LongName>Calibre-Web Catalog</LongName>
   <ShortName>Calibre-Web</ShortName>
   <Description>Calibre-Web Automated eBook Catalog</Description>
   <Developer>CrocodileStick</Developer>
   <Contact>https://github.com/crocodilestick/calibre-web-automated</Contact>
   <Url type="text/html" template="/search?query={searchTerms}"/>
   <Url type="application/atom+xml" template="/opds/search?query={searchTerms}"/>
   <SyndicationRight>open</SyndicationRight>
   <Language>en</Language>
   <OutputEncoding>UTF-8</OutputEncoding>
   <InputEncoding>UTF-8</InputEncoding>
</OpenSearchDescription>`
	w.Header().Set("Content-Type", "application/opensearchdescription+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(xmlStr))
}

func writeOPDS(w http.ResponseWriter, feed AtomFeed) {
	w.Header().Set("Content-Type", "application/atom+xml; profile=opds-catalog; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(feed)
}
