package sources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const userAgent = "Mozilla/5.0 (compatible; CalibreWebGo/1.0)"

func doReq(ctx context.Context, c *http.Client, method, u string, body io.Reader, hdr map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return b, nil
}

func getJSON(ctx context.Context, c *http.Client, u string, hdr map[string]string, out any) error {
	b, err := doReq(ctx, c, http.MethodGet, u, nil, hdr)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func first(s []string) string {
	if len(s) > 0 {
		return s[0]
	}
	return ""
}

// ---------------------------------------------------------------------------
// Open Library + Internet Archive (sem chave; apenas obras de acesso público)
// ---------------------------------------------------------------------------

type openLibrary struct{}

func (openLibrary) ID() string   { return "openlibrary" }
func (openLibrary) Name() string { return "Open Library / Internet Archive" }
func (openLibrary) Description() string {
	return "Obras de domínio público e de acesso livre. Não requer chave."
}
func (openLibrary) Fields() []Field                   { return nil }
func (openLibrary) Configured(map[string]string) bool { return true }

func (openLibrary) Search(ctx context.Context, c *http.Client, _ map[string]string, q string) ([]Result, error) {
	var r struct {
		Docs []struct {
			Title       string   `json:"title"`
			Authors     []string `json:"author_name"`
			Year        int      `json:"first_publish_year"`
			IA          []string `json:"ia"`
			CoverI      int      `json:"cover_i"`
			EbookAccess string   `json:"ebook_access"`
		} `json:"docs"`
	}
	u := "https://openlibrary.org/search.json?has_fulltext=true&limit=25&fields=title,author_name,first_publish_year,ia,cover_i,ebook_access&q=" + url.QueryEscape(q)
	if err := getJSON(ctx, c, u, nil, &r); err != nil {
		return nil, err
	}
	var out []Result
	for _, d := range r.Docs {
		if d.EbookAccess != "public" || len(d.IA) == 0 {
			continue
		}
		res := Result{Source: "openlibrary", ID: d.IA[0], Title: d.Title, Author: first(d.Authors)}
		if d.Year > 0 {
			res.Year = fmt.Sprint(d.Year)
		}
		if d.CoverI > 0 {
			res.Cover = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", d.CoverI)
		}
		out = append(out, res)
	}
	return out, nil
}

var iaID = regexp.MustCompile(`^[A-Za-z0-9._\-]+$`)

func (openLibrary) Resolve(ctx context.Context, c *http.Client, _ map[string]string, id string) (*Download, error) {
	if !iaID.MatchString(id) {
		return nil, errors.New("identificador inválido")
	}
	var m struct {
		Files []struct {
			Name   string `json:"name"`
			Format string `json:"format"`
		} `json:"files"`
	}
	if err := getJSON(ctx, c, "https://archive.org/metadata/"+id, nil, &m); err != nil {
		return nil, err
	}
	for _, want := range []string{".epub", ".pdf", ".mobi", ".azw3"} {
		for _, f := range m.Files {
			if strings.HasSuffix(strings.ToLower(f.Name), want) {
				return &Download{
					URL:      "https://archive.org/download/" + id + "/" + url.PathEscape(f.Name),
					Filename: f.Name,
				}, nil
			}
		}
	}
	return nil, errors.New("nenhum ficheiro EPUB/PDF/MOBI disponível para este item")
}

func (openLibrary) Test(ctx context.Context, c *http.Client, _ map[string]string) error {
	var r struct {
		Docs []struct{} `json:"docs"`
	}
	return getJSON(ctx, c, "https://openlibrary.org/search.json?limit=1&q=test", nil, &r)
}

// ---------------------------------------------------------------------------
// Project Gutenberg (via Gutendex; sem chave)
// ---------------------------------------------------------------------------

type gutenberg struct{}

func (gutenberg) ID() string   { return "gutenberg" }
func (gutenberg) Name() string { return "Project Gutenberg" }
func (gutenberg) Description() string {
	return "Mais de 70 mil livros de domínio público. Não requer chave."
}
func (gutenberg) Fields() []Field                   { return nil }
func (gutenberg) Configured(map[string]string) bool { return true }

type gutBook struct {
	ID      int `json:"id"`
	Title   string
	Authors []struct {
		Name string `json:"name"`
	} `json:"authors"`
	Formats map[string]string `json:"formats"`
}

func (gutenberg) Search(ctx context.Context, c *http.Client, _ map[string]string, q string) ([]Result, error) {
	var r struct {
		Results []gutBook `json:"results"`
	}
	if err := getJSON(ctx, c, "https://gutendex.com/books/?search="+url.QueryEscape(q), nil, &r); err != nil {
		return nil, err
	}
	var out []Result
	for _, b := range r.Results {
		res := Result{Source: "gutenberg", ID: fmt.Sprint(b.ID), Title: b.Title, Format: "epub", Cover: b.Formats["image/jpeg"]}
		if len(b.Authors) > 0 {
			res.Author = b.Authors[0].Name
		}
		out = append(out, res)
	}
	return out, nil
}

func (gutenberg) Resolve(ctx context.Context, c *http.Client, _ map[string]string, id string) (*Download, error) {
	var b gutBook
	if !regexp.MustCompile(`^\d+$`).MatchString(id) {
		return nil, errors.New("identificador inválido")
	}
	if err := getJSON(ctx, c, "https://gutendex.com/books/"+id, nil, &b); err != nil {
		return nil, err
	}
	exts := map[string]string{"application/epub+zip": ".epub", "application/x-mobipocket-ebook": ".mobi", "application/pdf": ".pdf"}
	for _, mt := range []string{"application/epub+zip", "application/x-mobipocket-ebook", "application/pdf"} {
		if u := b.Formats[mt]; u != "" {
			return &Download{URL: u, Filename: "gutenberg_" + id + exts[mt]}, nil
		}
	}
	return nil, errors.New("formato de livro eletrónico indisponível")
}

func (gutenberg) Test(ctx context.Context, c *http.Client, _ map[string]string) error {
	var r struct {
		Results []gutBook `json:"results"`
	}
	return getJSON(ctx, c, "https://gutendex.com/books/?search=test", nil, &r)
}

// ---------------------------------------------------------------------------
// Anna's Archive (chave de membro fornecida pelo utilizador)
// A pesquisa lê a página HTML (melhor esforço); o download usa a API oficial fast_download.
// ---------------------------------------------------------------------------

type annasArchive struct{}

func (annasArchive) ID() string   { return "annas" }
func (annasArchive) Name() string { return "Anna's Archive" }
func (annasArchive) Description() string {
	return "Requer a chave secreta de membro (API fast download) e um mirror ativo (os domínios mudam com frequência). " +
		"Não existe API de pesquisa: cole o MD5 ou o link /md5/... do livro para descarregar de forma fiável; " +
		"a pesquisa por texto é feita por leitura da página e pode falhar."
}
func (annasArchive) Fields() []Field {
	return []Field{
		{Key: "base_url", Label: "URL base de um mirror ativo (ex.: https://annas-archive.xx)"},
		{Key: "api_key", Label: "Chave de membro (secret key)", Secret: true},
	}
}
func (annasArchive) Configured(cfg map[string]string) bool {
	return cfg["api_key"] != "" && cfg["base_url"] != ""
}

var md5InText = regexp.MustCompile(`(?i)\b([0-9a-f]{32})\b`)

var (
	annaLink = regexp.MustCompile(`(?s)<a[^>]*href="/md5/([0-9a-f]{32})"[^>]*>(.*?)</a>`)
	htmlTag  = regexp.MustCompile(`(?s)<[^>]*>`)
	md5Re    = regexp.MustCompile(`^[0-9a-f]{32}$`)
)

func (annasArchive) Search(ctx context.Context, c *http.Client, cfg map[string]string, q string) ([]Result, error) {
	// Pesquisa por MD5 (ou link /md5/...): resultado direto, sem scraping.
	if m := md5InText.FindStringSubmatch(q); m != nil {
		md5 := strings.ToLower(m[1])
		return []Result{{Source: "annas", ID: md5, Title: "Ficheiro MD5 " + md5}}, nil
	}
	b, err := doReq(ctx, c, http.MethodGet, cfg["base_url"]+"/search?q="+url.QueryEscape(q), nil, nil)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []Result
	for _, m := range annaLink.FindAllStringSubmatch(string(b), -1) {
		if seen[m[1]] {
			continue
		}
		text := strings.Join(strings.Fields(html.UnescapeString(htmlTag.ReplaceAllString(m[2], " "))), " ")
		if len(text) < 3 {
			continue
		}
		seen[m[1]] = true
		out = append(out, Result{Source: "annas", ID: m[1], Title: text})
		if len(out) >= 25 {
			break
		}
	}
	if len(out) == 0 {
		return nil, errors.New("sem resultados legíveis (o mirror pode estar offline, protegido por captcha ou ter mudado de layout); cole o MD5 ou o link /md5/... do livro")
	}
	return out, nil
}

func (annasArchive) Resolve(ctx context.Context, c *http.Client, cfg map[string]string, id string) (*Download, error) {
	if !md5Re.MatchString(id) {
		return nil, errors.New("md5 inválido")
	}
	var r struct {
		URL   string `json:"download_url"`
		Error string `json:"error"`
	}
	u := cfg["base_url"] + "/dyn/api/fast_download.json?md5=" + id + "&key=" + url.QueryEscape(cfg["api_key"])
	b, err := doReq(ctx, c, http.MethodGet, u, nil, nil)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	if r.URL == "" {
		if r.Error == "" {
			r.Error = "resposta sem download_url"
		}
		return nil, errors.New("Anna's Archive: " + r.Error)
	}
	return &Download{URL: r.URL, Filename: id}, nil
}

func (annasArchive) Test(ctx context.Context, c *http.Client, cfg map[string]string) error {
	if cfg["base_url"] == "" {
		return errors.New("o URL base do mirror é obrigatório (ex: https://annas-archive.li)")
	}
	if cfg["api_key"] == "" {
		return errors.New("a chave de membro (secret key) é obrigatória")
	}
	u := strings.TrimRight(cfg["base_url"], "/") + "/dyn/api/fast_download.json?md5=00000000000000000000000000000000&key=" + url.QueryEscape(cfg["api_key"])
	b, err := doReq(ctx, c, http.MethodGet, u, nil, nil)
	if err != nil {
		return fmt.Errorf("mirror inacessível ou offline: %w", err)
	}
	var r struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(b, &r)
	if strings.Contains(strings.ToLower(r.Error), "invalid secret key") || strings.Contains(strings.ToLower(r.Error), "invalid key") {
		return errors.New("chave de membro inválida (verifique a Secret Key da sua conta Anna's Archive)")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Z-Library (API não oficial; credenciais remix_userid / remix_userkey do utilizador)
// ---------------------------------------------------------------------------

type zLibrary struct{}

func (zLibrary) ID() string   { return "zlibrary" }
func (zLibrary) Name() string { return "Z-Library" }
func (zLibrary) Description() string {
	return "API não oficial; requer o URL do seu domínio pessoal e os cookies remix_userid / remix_userkey da sua conta."
}
func (zLibrary) Fields() []Field {
	return []Field{
		{Key: "base_url", Label: "URL base (domínio pessoal)"},
		{Key: "user_id", Label: "remix_userid"},
		{Key: "user_key", Label: "remix_userkey", Secret: true},
	}
}
func (zLibrary) Configured(cfg map[string]string) bool {
	return cfg["base_url"] != "" && cfg["user_id"] != "" && cfg["user_key"] != ""
}

func zHeaders(cfg map[string]string) map[string]string {
	return map[string]string{
		"Cookie":        "remix_userid=" + cfg["user_id"] + "; remix_userkey=" + cfg["user_key"],
		"remix-userid":  cfg["user_id"],
		"remix-userkey": cfg["user_key"],
	}
}

func (zLibrary) Search(ctx context.Context, c *http.Client, cfg map[string]string, q string) ([]Result, error) {
	form := url.Values{"message": {q}, "limit": {"25"}}
	h := zHeaders(cfg)
	h["Content-Type"] = "application/x-www-form-urlencoded"
	b, err := doReq(ctx, c, http.MethodPost, cfg["base_url"]+"/eapi/book/search", strings.NewReader(form.Encode()), h)
	if err != nil {
		return nil, err
	}
	var r struct {
		Books []struct {
			ID     json.Number `json:"id"`
			Hash   string      `json:"hash"`
			Title  string      `json:"title"`
			Author string      `json:"author"`
			Year   json.Number `json:"year"`
			Ext    string      `json:"extension"`
			Size   string      `json:"filesizeString"`
			Cover  string      `json:"cover"`
		} `json:"books"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	var out []Result
	for _, bk := range r.Books {
		out = append(out, Result{
			Source: "zlibrary", ID: bk.ID.String() + "/" + bk.Hash, Title: bk.Title, Author: bk.Author,
			Year: bk.Year.String(), Format: bk.Ext, Size: bk.Size, Cover: bk.Cover,
		})
	}
	return out, nil
}

var zID = regexp.MustCompile(`^\d+/[0-9a-f]+$`)

func (zLibrary) Resolve(ctx context.Context, c *http.Client, cfg map[string]string, id string) (*Download, error) {
	if !zID.MatchString(id) {
		return nil, errors.New("identificador inválido")
	}
	var r struct {
		File struct {
			Link string `json:"downloadLink"`
			Ext  string `json:"extension"`
		} `json:"file"`
	}
	b, err := doReq(ctx, c, http.MethodGet, cfg["base_url"]+"/eapi/book/"+id+"/file", nil, zHeaders(cfg))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	if r.File.Link == "" {
		return nil, errors.New("Z-Library: sem link de download (limite diário atingido ou credenciais inválidas)")
	}
	name := strings.ReplaceAll(id, "/", "_")
	if r.File.Ext != "" {
		name += "." + r.File.Ext
	}
	return &Download{URL: r.File.Link, Filename: name, Header: http.Header{"Cookie": {zHeaders(cfg)["Cookie"]}}}, nil
}

func (zLibrary) Test(ctx context.Context, c *http.Client, cfg map[string]string) error {
	if cfg["base_url"] == "" || cfg["user_id"] == "" || cfg["user_key"] == "" {
		return errors.New("URL base, remix_userid e remix_userkey são todos obrigatórios")
	}
	form := url.Values{"message": {"test"}, "limit": {"1"}}
	h := zHeaders(cfg)
	h["Content-Type"] = "application/x-www-form-urlencoded"
	b, err := doReq(ctx, c, http.MethodPost, strings.TrimRight(cfg["base_url"], "/")+"/eapi/book/search", strings.NewReader(form.Encode()), h)
	if err != nil {
		return fmt.Errorf("falha ao contactar domínio Z-Library: %w", err)
	}
	var r struct {
		Success int    `json:"success"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(b, &r)
	if r.Error != "" || r.Message != "" {
		msg := r.Error
		if msg == "" {
			msg = r.Message
		}
		return errors.New("Z-Library: " + msg)
	}
	return nil
}
