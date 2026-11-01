package sources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Testes contra serviços reais; correr com: LIVE_SOURCES=1 go test ./internal/sources -run Live -v
func TestLiveFreeSources(t *testing.T) {
	if os.Getenv("LIVE_SOURCES") == "" {
		t.Skip("defina LIVE_SOURCES=1 para testar contra serviços reais")
	}
	m := NewManager(filepath.Join(t.TempDir(), "cfg.json"))
	for _, tc := range []struct{ src, q string }{
		{"gutenberg", "pride and prejudice"},
		{"openlibrary", "pride and prejudice"},
	} {
		res := m.Search(context.Background(), tc.q, tc.src)
		if e := res.Errors[tc.src]; e != "" {
			t.Fatalf("%s search: %s", tc.src, e)
		}
		if len(res.Results) == 0 {
			t.Fatalf("%s: sem resultados", tc.src)
		}
		var path string
		var err error
		for _, r := range res.Results { // alguns itens do Open Library podem não ter EPUB/PDF
			dir := t.TempDir()
			path, err = m.DownloadToDir(context.Background(), tc.src, r.ID, dir)
			if err == nil {
				break
			}
			t.Logf("%s/%s: %v", tc.src, r.ID, err)
		}
		if err != nil {
			t.Fatalf("%s: nenhum download funcionou: %v", tc.src, err)
		}
		st, _ := os.Stat(path)
		t.Logf("%s -> %s (%d bytes)", tc.src, filepath.Base(path), st.Size())
		if st.Size() < 10_000 {
			t.Fatalf("%s: ficheiro demasiado pequeno", tc.src)
		}
	}
}

func TestAnnasFastDownloadFlow(t *testing.T) {
	const md5 = "d6e1dc51a50726f00ec438af21952a45"
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dyn/api/fast_download.json":
			if r.URL.Query().Get("md5") != md5 || r.URL.Query().Get("key") != "k" {
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid secret key"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"download_url": srv.URL + "/file/book.epub"})
		case "/file/book.epub":
			w.Header().Set("Content-Type", "application/epub+zip")
			_, _ = w.Write([]byte(strings.Repeat("x", 100)))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	m := NewManager(filepath.Join(t.TempDir(), "cfg.json"))
	if err := m.Configure("annas", map[string]string{"base_url": srv.URL, "api_key": "k"}, true); err != nil {
		t.Fatal(err)
	}
	res := m.Search(context.Background(), "https://x.example/md5/"+md5, "annas")
	if len(res.Results) != 1 || res.Results[0].ID != md5 {
		t.Fatalf("pesquisa por md5: %+v", res)
	}
	dir := t.TempDir()
	p, err := m.DownloadToDir(context.Background(), "annas", md5, dir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(p) != ".epub" {
		t.Fatalf("extensão esperada .epub: %s", p)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal(err)
	}
	// chave errada devolve erro claro
	_ = m.Configure("annas", map[string]string{"api_key": "bad"}, true)
	if _, err := m.DownloadToDir(context.Background(), "annas", md5, dir); err == nil || !strings.Contains(err.Error(), "Invalid secret key") {
		t.Fatalf("esperava erro de chave, obtive %v", err)
	}
}

func TestZLibraryFlow(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Cookie"), "remix_userid=1") {
			http.Error(w, "no auth", 401)
			return
		}
		switch {
		case r.URL.Path == "/eapi/book/search" && r.Method == "POST":
			_ = r.ParseForm()
			if r.Form.Get("message") != "dune" {
				http.Error(w, "bad", 400)
				return
			}
			_, _ = w.Write([]byte(`{"success":1,"books":[{"id":"123","hash":"abc123","title":"Dune","author":"Frank Herbert","year":"1965","extension":"epub","filesizeString":"1 MB","cover":""}]}`))
		case r.URL.Path == "/eapi/book/123/abc123/file":
			_, _ = w.Write([]byte(`{"success":1,"file":{"downloadLink":"` + srv.URL + `/dl/x","extension":"epub"}}`))
		case r.URL.Path == "/dl/x":
			w.Header().Set("Content-Type", "application/epub+zip")
			_, _ = w.Write([]byte("PK-data"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	m := NewManager(filepath.Join(t.TempDir(), "cfg.json"))
	if err := m.Configure("zlibrary", map[string]string{"base_url": srv.URL, "user_id": "1", "user_key": "k"}, true); err != nil {
		t.Fatal(err)
	}
	res := m.Search(context.Background(), "dune", "zlibrary")
	if len(res.Results) != 1 || res.Results[0].ID != "123/abc123" {
		t.Fatalf("pesquisa: %+v", res)
	}
	p, err := m.DownloadToDir(context.Background(), "zlibrary", "123/abc123", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(p, ".epub") {
		t.Fatalf("extensão: %s", p)
	}
}

func TestSecretsMaskedAndPersisted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.json")
	m := NewManager(path)
	_ = m.Configure("annas", map[string]string{"base_url": "https://m.example", "api_key": "secret"}, true)
	for _, i := range m.List() {
		if i.ID == "annas" && i.Values["api_key"] != "********" {
			t.Fatalf("segredo não mascarado: %v", i.Values)
		}
	}
	_ = m.Configure("annas", map[string]string{"api_key": "********"}, true) // mantém
	if NewManager(path).cfg["annas"]["api_key"] != "secret" {
		t.Fatal("segredo perdido")
	}
}
