// Package sources define um sistema de fontes de livros (providers) pesquisáveis e descarregáveis.
// Cada fonte implementa Provider; as credenciais/chaves são fornecidas pelo utilizador e
// guardadas em disco (ficheiro JSON com permissões restritas).
package sources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Result é um resultado de pesquisa normalizado.
type Result struct {
	Source string `json:"source"`
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author,omitempty"`
	Year   string `json:"year,omitempty"`
	Format string `json:"format,omitempty"`
	Size   string `json:"size,omitempty"`
	Cover  string `json:"cover,omitempty"`
}

// Download descreve onde obter o ficheiro de um resultado.
type Download struct {
	URL      string
	Filename string // opcional; se vazio é deduzido dos cabeçalhos/URL
	Header   http.Header
}

// Field descreve um campo de configuração apresentado na UI.
type Field struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Secret   bool   `json:"secret"`
	Optional bool   `json:"optional"`
	Default  string `json:"default,omitempty"`
}

// Provider é uma fonte de livros.
type Provider interface {
	ID() string
	Name() string
	Description() string
	Fields() []Field
	Configured(cfg map[string]string) bool
	Search(ctx context.Context, c *http.Client, cfg map[string]string, query string) ([]Result, error)
	Resolve(ctx context.Context, c *http.Client, cfg map[string]string, id string) (*Download, error)
}

// Info é o estado público de uma fonte (sem segredos).
type Info struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Fields      []Field           `json:"fields"`
	Configured  bool              `json:"configured"`
	Enabled     bool              `json:"enabled"`
	Values      map[string]string `json:"values"` // segredos mascarados
}

// Manager guarda a configuração e coordena pesquisas/downloads.
type Manager struct {
	mu        sync.RWMutex
	path      string
	providers map[string]Provider
	cfg       map[string]map[string]string
	client    *http.Client
}

// NewManager cria o gestor; configPath é o ficheiro JSON onde a configuração é guardada.
func NewManager(configPath string) *Manager {
	m := &Manager{
		path:      configPath,
		providers: map[string]Provider{},
		cfg:       map[string]map[string]string{},
		client:    &http.Client{Timeout: 60 * time.Second},
	}
	for _, p := range []Provider{openLibrary{}, gutenberg{}, annasArchive{}, zLibrary{}} {
		m.providers[p.ID()] = p
	}
	if b, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(b, &m.cfg)
	}
	return m
}

func (m *Manager) sortedProviders() []Provider {
	out := make([]Provider, 0, len(m.providers))
	for _, p := range m.providers {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}

func (m *Manager) values(p Provider) map[string]string {
	v := map[string]string{}
	for _, f := range p.Fields() {
		if f.Default != "" {
			v[f.Key] = f.Default
		}
	}
	for k, s := range m.cfg[p.ID()] {
		v[k] = s
	}
	return v
}

func (m *Manager) enabled(p Provider) bool {
	return m.cfg[p.ID()]["enabled"] != "false" && p.Configured(m.values(p))
}

// List devolve o estado de todas as fontes, com segredos mascarados.
func (m *Manager) List() []Info {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Info
	for _, p := range m.sortedProviders() {
		vals := m.values(p)
		masked := map[string]string{}
		for _, f := range p.Fields() {
			if f.Secret && vals[f.Key] != "" {
				masked[f.Key] = "********"
			} else {
				masked[f.Key] = vals[f.Key]
			}
		}
		out = append(out, Info{
			ID: p.ID(), Name: p.Name(), Description: p.Description(), Fields: p.Fields(),
			Configured: p.Configured(vals), Enabled: m.enabled(p), Values: masked,
		})
	}
	return out
}

// Configure atualiza a configuração de uma fonte. Valores "********" mantêm o segredo atual.
func (m *Manager) Configure(id string, values map[string]string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.providers[id]
	if !ok {
		return fmt.Errorf("fonte desconhecida: %s", id)
	}
	cur := m.cfg[id]
	if cur == nil {
		cur = map[string]string{}
	}
	for _, f := range p.Fields() {
		v, present := values[f.Key]
		if !present || v == "********" {
			continue
		}
		v = strings.TrimSpace(v)
		if f.Key == "base_url" && v != "" {
			u, err := url.Parse(v)
			if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
				return errors.New("base_url inválido")
			}
			v = strings.TrimRight(v, "/")
		}
		cur[f.Key] = v
	}
	cur["enabled"] = fmt.Sprint(enabled)
	m.cfg[id] = cur
	b, err := json.MarshalIndent(m.cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, b, 0o600)
}

// SearchResult agrega resultados e erros por fonte.
type SearchResult struct {
	Results []Result          `json:"results"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// Search pesquisa em paralelo nas fontes ativas (ou apenas em onlySource, se indicado).
func (m *Manager) Search(ctx context.Context, query, onlySource string) SearchResult {
	m.mu.RLock()
	type job struct {
		p   Provider
		cfg map[string]string
	}
	var jobs []job
	for _, p := range m.sortedProviders() {
		if onlySource != "" && onlySource != p.ID() {
			continue
		}
		if m.enabled(p) {
			jobs = append(jobs, job{p, m.values(p)})
		}
	}
	m.mu.RUnlock()

	res := SearchResult{Results: []Result{}, Errors: map[string]string{}}
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, j := range jobs {
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			cctx, cancel := context.WithTimeout(ctx, 25*time.Second)
			defer cancel()
			r, err := j.p.Search(cctx, m.client, j.cfg, query)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				res.Errors[j.p.ID()] = err.Error()
				return
			}
			res.Results = append(res.Results, r...)
		}(j)
	}
	wg.Wait()
	sort.SliceStable(res.Results, func(i, k int) bool { return res.Results[i].Source < res.Results[k].Source })
	return res
}

var unsafeName = regexp.MustCompile(`[^\w.\- ()\[\]]+`)

// DownloadToDir descarrega o livro para destDir e devolve o caminho final.
func (m *Manager) DownloadToDir(ctx context.Context, source, id, destDir string) (string, error) {
	m.mu.RLock()
	p, ok := m.providers[source]
	var cfg map[string]string
	if ok {
		cfg = m.values(p)
	}
	enabled := ok && m.enabled(p)
	m.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("fonte desconhecida: %s", source)
	}
	if !enabled {
		return "", fmt.Errorf("fonte %s não está configurada/ativa", p.Name())
	}

	d, err := p.Resolve(ctx, m.client, cfg, id)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(d.URL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return "", errors.New("URL de download inválido")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.URL, nil)
	if err != nil {
		return "", err
	}
	for k, vs := range d.Header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("User-Agent", userAgent)
	dl := &http.Client{Timeout: 15 * time.Minute}
	resp, err := dl.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download falhou: HTTP %d", resp.StatusCode)
	}
	if strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return "", errors.New("a fonte devolveu uma página HTML em vez de um ficheiro (chave inválida, limite diário ou captcha)")
	}

	name := d.Filename
	if _, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition")); err == nil && params["filename"] != "" {
		name = params["filename"]
	}
	if name == "" {
		name = filepath.Base(u.Path)
	}
	name = strings.TrimSpace(unsafeName.ReplaceAllString(filepath.Base(name), "_"))
	if name == "" || name == "." || name == "_" {
		name = "book_" + id
	}

	tmp, err := os.CreateTemp("", "cwa-dl-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, io.LimitReader(resp.Body, 2<<30)); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}
	dest := filepath.Join(destDir, name)
	// Copia para o destino com nome temporário e renomeia no fim (o watcher só vê o ficheiro completo).
	part := dest + ".part"
	if err := copyFile(tmp.Name(), part); err != nil {
		return "", err
	}
	if err := os.Rename(part, dest); err != nil {
		return "", err
	}
	return dest, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
