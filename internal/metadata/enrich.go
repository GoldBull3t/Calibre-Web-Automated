package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CandidateCover representa uma opção de capa encontrada online em alta resolução.
type CandidateCover struct {
	ID       string `json:"id"`
	URL      string `json:"url"`       // Capa em Alta Definição (HD)
	ThumbURL string `json:"thumb_url"` // Miniatura para pré-visualização rápida
}

// EnrichmentCandidate agrega os metadados enriquecidos prontos para visualização e aplicação.
type EnrichmentCandidate struct {
	Provider    string            `json:"provider"`
	WorkKey     string            `json:"work_key"`
	Title       string            `json:"title"`
	Authors     []string          `json:"authors"`
	Description string            `json:"description"`
	Covers      []CandidateCover  `json:"covers"`
	Publisher   string            `json:"publisher"`
	PublishDate string            `json:"publish_date"`
	Tags        []string          `json:"tags"`
	Rating      float64           `json:"rating"`
	Identifiers map[string]string `json:"identifiers"`
}

// FetchEnrichmentCandidates pesquisa em múltiplos provedores online (OpenLibrary, Google Books)
// e devolve candidatos com capas HD e sinopses completas.
func FetchEnrichmentCandidates(ctx context.Context, query, title, author string) ([]EnrichmentCandidate, error) {
	searchQuery := strings.TrimSpace(query)
	if searchQuery == "" {
		cleanedTitle := CleanBookTitle(title)
		cleanedAuthor := CleanAuthorName(author)
		searchQuery = strings.TrimSpace(cleanedTitle + " " + cleanedAuthor)
	}
	if searchQuery == "" {
		return nil, nil
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []EnrichmentCandidate

	// 1. OpenLibrary Search + Work Enrichment
	wg.Add(1)
	go func() {
		defer wg.Done()
		olCandidates, err := searchOpenLibraryEnriched(ctxTimeout, searchQuery)
		if err == nil && len(olCandidates) > 0 {
			mu.Lock()
			results = append(results, olCandidates...)
			mu.Unlock()
		}
	}()

	// 2. Google Books Search (se disponível)
	wg.Add(1)
	go func() {
		defer wg.Done()
		gbCandidates, err := SearchGoogleBooks(ctxTimeout, searchQuery)
		if err == nil && len(gbCandidates) > 0 {
			var converted []EnrichmentCandidate
			for _, gb := range gbCandidates {
				var covers []CandidateCover
				if gb.Cover != "" {
					hdCover := gb.Cover
					// Google Books: zoom=1 para zoom=2/3 remove limitações de resolução
					hdCover = strings.Replace(hdCover, "zoom=1", "zoom=2", 1)
					hdCover = strings.Replace(hdCover, "&edge=curl", "", 1)
					covers = append(covers, CandidateCover{
						ID:       gb.ID,
						URL:      hdCover,
						ThumbURL: gb.Cover,
					})
				}
				converted = append(converted, EnrichmentCandidate{
					Provider:    "google",
					WorkKey:     gb.ID,
					Title:       gb.Title,
					Authors:     gb.Authors,
					Description: gb.Description,
					Covers:      covers,
					Publisher:   gb.Publisher,
					PublishDate: gb.PublishedDate,
					Tags:        gb.Tags,
					Rating:      gb.Rating,
					Identifiers: gb.Identifiers,
				})
			}
			mu.Lock()
			results = append(results, converted...)
			mu.Unlock()
		}
	}()

	wg.Wait()
	return results, nil
}

// searchOpenLibraryEnriched faz a pesquisa e busca detalhada do work JSON para obter todas as capas e sinopse.
func searchOpenLibraryEnriched(ctx context.Context, query string) ([]EnrichmentCandidate, error) {
	endpoint := "https://openlibrary.org/search.json?q=" + url.QueryEscape(query) + "&limit=4"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Calibre-Web-Automated-Go/2.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open library status: %d", resp.StatusCode)
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

	candidates := make([]EnrichmentCandidate, 0, len(data.Docs))

	for _, doc := range data.Docs {
		if doc.Title == "" || doc.Key == "" {
			continue
		}

		cand := EnrichmentCandidate{
			Provider: "openlibrary",
			WorkKey:  doc.Key,
			Title:    CleanBookTitle(doc.Title),
			Authors:  cleanAuthorSlice(doc.AuthorName),
			Tags:     filterMeaningfulTags(doc.Subject),
			Identifiers: map[string]string{
				"openlibrary": strings.TrimPrefix(doc.Key, "/works/"),
			},
		}

		if doc.FirstPublish > 0 {
			cand.PublishDate = fmt.Sprintf("%d", doc.FirstPublish)
		}
		if len(doc.Publisher) > 0 {
			cand.Publisher = doc.Publisher[0]
		}
		if len(doc.ISBN) > 0 {
			cand.Identifiers["isbn"] = doc.ISBN[0]
		}

		// Obter Work Details (covers[] e description completa)
		workDetails, _ := fetchOpenLibraryWorkDetails(ctx, doc.Key)
		if workDetails != nil {
			if workDetails.Description != "" {
				cand.Description = workDetails.Description
			}
			if len(workDetails.Tags) > 0 {
				cand.Tags = mergeTags(cand.Tags, workDetails.Tags)
			}
			for k, v := range workDetails.Identifiers {
				cand.Identifiers[k] = v
			}

			// Capas HD encontradas
			for _, covID := range workDetails.Covers {
				cand.Covers = append(cand.Covers, CandidateCover{
					ID:       strconv.FormatInt(covID, 10),
					URL:      fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", covID),
					ThumbURL: fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", covID),
				})
			}
		}

		// Fallback para capa do doc se não houver no work
		if len(cand.Covers) == 0 && doc.CoverI > 0 {
			cand.Covers = append(cand.Covers, CandidateCover{
				ID:       strconv.FormatInt(doc.CoverI, 10),
				URL:      fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", doc.CoverI),
				ThumbURL: fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", doc.CoverI),
			})
		}

		candidates = append(candidates, cand)
	}

	return candidates, nil
}

type olWorkDetail struct {
	Description string
	Covers      []int64
	Tags        []string
	Identifiers map[string]string
}

func fetchOpenLibraryWorkDetails(ctx context.Context, workKey string) (*olWorkDetail, error) {
	if !strings.HasPrefix(workKey, "/") {
		workKey = "/" + workKey
	}
	urlStr := "https://openlibrary.org" + workKey + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Calibre-Web-Automated-Go/2.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("work status %d", resp.StatusCode)
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	detail := &olWorkDetail{
		Identifiers: make(map[string]string),
	}

	// 1. Extrair Description (string ou {"value": "..."})
	if dVal, ok := raw["description"]; ok && dVal != nil {
		switch v := dVal.(type) {
		case string:
			detail.Description = CleanGutenbergText(v)
		case map[string]any:
			if strVal, hasStr := v["value"].(string); hasStr {
				detail.Description = CleanGutenbergText(strVal)
			}
		}
	}

	// 2. Extrair Covers
	if covList, ok := raw["covers"].([]any); ok {
		for _, c := range covList {
			switch cv := c.(type) {
			case float64:
				if int64(cv) > 0 {
					detail.Covers = append(detail.Covers, int64(cv))
				}
			}
		}
	}

	// 3. Extrair Subjects
	if subjList, ok := raw["subjects"].([]any); ok {
		var subjs []string
		for _, s := range subjList {
			if sStr, isStr := s.(string); isStr {
				subjs = append(subjs, sStr)
			}
		}
		detail.Tags = filterMeaningfulTags(subjs)
	}

	// 4. Identifiers (Goodreads, LibraryThing, Wikidata)
	if idMap, ok := raw["identifiers"].(map[string]any); ok {
		for k, v := range idMap {
			if arr, isArr := v.([]any); isArr && len(arr) > 0 {
				if firstStr, okStr := arr[0].(string); okStr {
					detail.Identifiers[k] = firstStr
				}
			}
		}
	}

	return detail, nil
}

// CleanAuthorName sanitiza nomes de autores com vírgulas invertidas, séculos ou artefatos.
func CleanAuthorName(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}

	// Casos específicos conhecidos do Gutenberg
	lower := strings.ToLower(s)
	if strings.Contains(lower, "sunzi") || strings.Contains(lower, "sun tzu") {
		return "Sun Tzu"
	}

	// Remove menções de datas/séculos como "active 6th century B.C."
	reDate := regexp.MustCompile(`(?i)(active|fl\.|circa|c\.)\s+[^,]+`)
	s = reDate.ReplaceAllString(s, "")

	// Remove anos de nascimento/morte: (1797-1851) ou , 1797-1851
	reYears := regexp.MustCompile(`[,\s]*\(?\d{3,4}\s*[-–—]\s*\d{3,4}\)?`)
	s = reYears.ReplaceAllString(s, "")

	// Inverte "Apelido, Nome" -> "Nome Apelido"
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		if len(parts) == 2 {
			p1 := strings.TrimSpace(parts[0])
			p2 := strings.TrimSpace(parts[1])
			// Apenas inverte se a segunda parte parecer um primeiro nome
			if len(p2) > 0 && !strings.ContainsAny(p2, "0123456789") {
				s = p2 + " " + p1
			}
		}
	}

	// Remove parênteses adicionais e espaços duplos
	reParen := regexp.MustCompile(`\([^)]*\)`)
	s = reParen.ReplaceAllString(s, "")

	s = strings.Join(strings.Fields(s), " ")
	return strings.Trim(s, " ,-")
}

// CleanBookTitle remove sufixos chatos de edições ou pontuação excessiva.
func CleanBookTitle(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}

	// Remove "(eBook)", "[Project Gutenberg]", etc.
	reBrackets := regexp.MustCompile(`(?i)\s*[\(\[](e-?book|project gutenberg|illustrated|edition)[^\)\]]*[\)\]]`)
	s = reBrackets.ReplaceAllString(s, "")

	// Remove números de sufixo no título "(2)", "(11)" comuns de caminhos
	reNum := regexp.MustCompile(`\s*\(\d+\)$`)
	s = reNum.ReplaceAllString(s, "")

	s = strings.Join(strings.Fields(s), " ")
	return s
}

// CleanGutenbergText limpa preâmbulos, avisos legais e transcrições de ebooks do Project Gutenberg.
func CleanGutenbergText(text string) string {
	if text == "" {
		return ""
	}

	// Remove marcas de início e fim do Project Gutenberg
	reStart := regexp.MustCompile(`(?is)^\s*\*\*\*\s*START OF (THE|THIS) PROJECT GUTENBERG EBOOK.*?\*\*\*`)
	text = reStart.ReplaceAllString(text, "")

	reEnd := regexp.MustCompile(`(?is)\*\*\*\s*END OF (THE|THIS) PROJECT GUTENBERG EBOOK.*$`)
	text = reEnd.ReplaceAllString(text, "")

	// Remove cabeçalhos "The Project Gutenberg eBook of..."
	reHeader := regexp.MustCompile(`(?im)^The Project Gutenberg eBook of .*?$`)
	text = reHeader.ReplaceAllString(text, "")

	return strings.TrimSpace(text)
}

func cleanAuthorSlice(authors []string) []string {
	var cleaned []string
	seen := make(map[string]bool)
	for _, a := range authors {
		c := CleanAuthorName(a)
		if c != "" && !seen[c] {
			seen[c] = true
			cleaned = append(cleaned, c)
		}
	}
	return cleaned
}

func filterMeaningfulTags(tags []string) []string {
	var filtered []string
	banned := map[string]bool{
		"fiction": true, "general": true, "accessible book": true,
		"protected daisy": true, "in library": true, "open library": true,
		"classic": true, "novel": true, "novela": true,
	}

	seen := make(map[string]bool)
	for _, t := range tags {
		t = strings.TrimSpace(t)
		lower := strings.ToLower(t)
		if len(t) < 3 || banned[lower] || seen[lower] {
			continue
		}
		seen[lower] = true
		filtered = append(filtered, t)
		if len(filtered) >= 6 {
			break
		}
	}
	return filtered
}

func mergeTags(t1, t2 []string) []string {
	seen := make(map[string]bool)
	var merged []string
	for _, t := range t1 {
		l := strings.ToLower(t)
		if !seen[l] {
			seen[l] = true
			merged = append(merged, t)
		}
	}
	for _, t := range t2 {
		l := strings.ToLower(t)
		if !seen[l] {
			seen[l] = true
			merged = append(merged, t)
		}
		if len(merged) >= 8 {
			break
		}
	}
	return merged
}

