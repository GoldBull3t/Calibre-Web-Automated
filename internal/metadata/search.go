package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OnlineBookMeta contém metadados enriquecidos obtidos de provedores online.
type OnlineBookMeta struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Authors       []string          `json:"authors"`
	Description   string            `json:"description"`
	Cover         string            `json:"cover"`
	Publisher     string            `json:"publisher"`
	PublishedDate string            `json:"publishedDate"`
	Rating        float64           `json:"rating"`
	Tags          []string          `json:"tags"`
	Identifiers   map[string]string `json:"identifiers"`
	Source        string            `json:"source"`
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// SearchGoogleBooks pesquisa livros na API pública do Google Books.
func SearchGoogleBooks(ctx context.Context, query string) ([]OnlineBookMeta, error) {
	if strings.TrimSpace(query) == "" {
		return nil, nil
	}

	endpoint := "https://www.googleapis.com/books/v1/volumes?q=" + url.QueryEscape(query) + "&maxResults=10"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Calibre-Web-Automated-Go/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google books API returned status %d", resp.StatusCode)
	}

	var data struct {
		Items []struct {
			ID         string `json:"id"`
			VolumeInfo struct {
				Title         string   `json:"title"`
				Authors       []string `json:"authors"`
				Publisher     string   `json:"publisher"`
				PublishedDate string   `json:"publishedDate"`
				Description   string   `json:"description"`
				AverageRating float64  `json:"averageRating"`
				Categories    []string `json:"categories"`
				ImageLinks    struct {
					Thumbnail      string `json:"thumbnail"`
					SmallThumbnail string `json:"smallThumbnail"`
				} `json:"imageLinks"`
				IndustryIdentifiers []struct {
					Type       string `json:"type"`
					Identifier string `json:"identifier"`
				} `json:"industryIdentifiers"`
			} `json:"volumeInfo"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	results := make([]OnlineBookMeta, 0, len(data.Items))
	for _, item := range data.Items {
		info := item.VolumeInfo
		if info.Title == "" {
			continue
		}

		cover := info.ImageLinks.Thumbnail
		if cover == "" {
			cover = info.ImageLinks.SmallThumbnail
		}
		if cover != "" && strings.HasPrefix(cover, "http://") {
			cover = "https://" + strings.TrimPrefix(cover, "http://")
		}

		identifiers := map[string]string{
			"google": item.ID,
		}
		for _, id := range info.IndustryIdentifiers {
			if strings.EqualFold(id.Type, "ISBN_13") || strings.EqualFold(id.Type, "ISBN_10") {
				identifiers["isbn"] = id.Identifier
			}
		}

		results = append(results, OnlineBookMeta{
			ID:            item.ID,
			Title:         info.Title,
			Authors:       info.Authors,
			Description:   info.Description,
			Cover:         cover,
			Publisher:     info.Publisher,
			PublishedDate: info.PublishedDate,
			Rating:        info.AverageRating,
			Tags:          info.Categories,
			Identifiers:   identifiers,
			Source:        "google",
		})
	}

	return results, nil
}

// SearchOpenLibrary pesquisa livros na API pública do Open Library.
func SearchOpenLibrary(ctx context.Context, query string) ([]OnlineBookMeta, error) {
	if strings.TrimSpace(query) == "" {
		return nil, nil
	}

	endpoint := "https://openlibrary.org/search.json?q=" + url.QueryEscape(query) + "&limit=10"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Calibre-Web-Automated-Go/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open library API returned status %d", resp.StatusCode)
	}

	var data struct {
		Docs []struct {
			Key          string   `json:"key"`
			Title        string   `json:"title"`
			AuthorName   []string `json:"author_name"`
			Publisher    []string `json:"publisher"`
			FirstPublish int      `json:"first_publish_year"`
			CoverI       int64    `json:"cover_i"`
			ISBN         []string `json:"isbn"`
			Subject      []string `json:"subject"`
		} `json:"docs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	results := make([]OnlineBookMeta, 0, len(data.Docs))
	for _, doc := range data.Docs {
		if doc.Title == "" {
			continue
		}

		var cover string
		if doc.CoverI > 0 {
			cover = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", doc.CoverI)
		}

		identifiers := map[string]string{
			"openlibrary": strings.TrimPrefix(doc.Key, "/works/"),
		}
		if len(doc.ISBN) > 0 {
			identifiers["isbn"] = doc.ISBN[0]
		}

		var pubDate string
		if doc.FirstPublish > 0 {
			pubDate = fmt.Sprintf("%d", doc.FirstPublish)
		}

		var publisher string
		if len(doc.Publisher) > 0 {
			publisher = doc.Publisher[0]
		}

		tags := doc.Subject
		if len(tags) > 5 {
			tags = tags[:5]
		}

		results = append(results, OnlineBookMeta{
			ID:            doc.Key,
			Title:         doc.Title,
			Authors:       doc.AuthorName,
			Cover:         cover,
			Publisher:     publisher,
			PublishedDate: pubDate,
			Tags:          tags,
			Identifiers:   identifiers,
			Source:        "openlibrary",
		})
	}

	return results, nil
}

// SearchOnline agrega resultados de múltiplos provedores de metadados em paralelo.
func SearchOnline(ctx context.Context, query string) []OnlineBookMeta {
	ctxTimeout, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	type res struct {
		items []OnlineBookMeta
	}
	ch := make(chan res, 2)

	go func() {
		items, _ := SearchGoogleBooks(ctxTimeout, query)
		ch <- res{items: items}
	}()

	go func() {
		items, _ := SearchOpenLibrary(ctxTimeout, query)
		ch <- res{items: items}
	}()

	var all []OnlineBookMeta
	for i := 0; i < 2; i++ {
		r := <-ch
		if len(r.items) > 0 {
			all = append(all, r.items...)
		}
	}
	return all
}

