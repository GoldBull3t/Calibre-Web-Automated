package mam

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMAMClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") != "mam_id=test_cookie_123" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path == "/json/load.php" && r.URL.Query().Get("tor[text]") != "" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"data": [
					{
						"id": 12345,
						"title": "Dune",
						"author_info": "Frank Herbert",
						"filetype": "epub",
						"size": "2.4 MB",
						"seeders": 42,
						"leechers": 1,
						"vip": 0,
						"free": 1
					}
				]
			}`))
			return
		}

		if r.URL.Path == "/json/load.php" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"username": "reader42",
				"uid": 9999,
				"ratio": 4.52,
				"uploaded": "150 GB",
				"downloaded": "33.2 GB",
				"seedbonus": 12500.5,
				"vip": true
			}`))
			return
		}

		if r.URL.Path == "/tor/download.php" {
			w.Header().Set("Content-Type", "application/x-bittorrent")
			_, _ = w.Write([]byte("mock_torrent_content"))
			return
		}

		http.NotFound(w, r)
	}))
	defer ts.Close()

	c := NewClient("test_cookie_123")
	c.baseURL = ts.URL

	ctx := context.Background()

	// 1. Testa Account Info
	info, err := c.GetAccountInfo(ctx)
	if err != nil {
		t.Fatalf("erro ao obter account info: %v", err)
	}
	if info.Username != "reader42" || info.Ratio != 4.52 || !info.IsVIP {
		t.Errorf("dados de conta incorretos: %+v", info)
	}

	// 2. Testa Search
	items, err := c.Search(ctx, "Dune", 10)
	if err != nil {
		t.Fatalf("erro na pesquisa MAM: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Dune" || !items[0].Freeleech {
		t.Errorf("resultados de pesquisa incorretos: %+v", items)
	}

	// 3. Testa Download
	tmpDir, err := os.MkdirTemp("", "mam_test_*")
	if err != nil {
		t.Fatalf("erro ao criar pasta temporária: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	downloadPath, err := c.DownloadTorrentToIngest(ctx, 12345, tmpDir)
	if err != nil {
		t.Fatalf("erro ao descarregar torrent: %v", err)
	}
	content, err := os.ReadFile(downloadPath)
	if err != nil || string(content) != "mock_torrent_content" {
		t.Errorf("conteúdo do ficheiro descarregado inválido")
	}
}

