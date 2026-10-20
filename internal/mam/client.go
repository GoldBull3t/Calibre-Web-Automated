package mam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	mamID      string
	httpClient *http.Client
	baseURL    string
}

type AccountInfo struct {
	Success      bool    `json:"success"`
	Username     string  `json:"username"`
	UserID       int64   `json:"user_id"`
	Ratio        float64 `json:"ratio"`
	Uploaded     string  `json:"uploaded"`
	Downloaded   string  `json:"downloaded"`
	BonusPoints  float64 `json:"bonus_points"`
	IsVIP        bool    `json:"is_vip"`
	VIPExpires   string  `json:"vip_expires"`
	RawResponse  string  `json:"-"`
}

type TorrentItem struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Narrator  string `json:"narrator,omitempty"`
	Series    string `json:"series,omitempty"`
	FileType  string `json:"file_type"`
	Size      string `json:"size"`
	SizeBytes int64  `json:"size_bytes"`
	Seeders   int    `json:"seeders"`
	Leechers  int    `json:"leechers"`
	VIP       bool   `json:"vip"`
	Freeleech bool   `json:"freeleech"`
}

func NewClient(mamID string) *Client {
	return &Client{
		mamID:   strings.TrimSpace(mamID),
		baseURL: "https://www.myanonamouse.net",
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) SetMamID(id string) {
	c.mamID = strings.TrimSpace(id)
}

func (c *Client) HasAuth() bool {
	return c.mamID != ""
}

func (c *Client) GetAccountInfo(ctx context.Context) (*AccountInfo, error) {
	if !c.HasAuth() {
		return nil, fmt.Errorf("mam_id não configurado")
	}

	reqURL := fmt.Sprintf("%s/json/load.php", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Calibre-Web-Automated/2026.5")
	req.Header.Set("Cookie", fmt.Sprintf("mam_id=%s", c.mamID))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na ligação à API do MAM: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("autenticação inválida ou cookie mam_id expirado")
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("resposta inesperada do MAM: %w", err)
	}

	info := &AccountInfo{
		Success:     true,
		RawResponse: string(body),
	}

	if u, ok := raw["username"].(string); ok {
		info.Username = u
	}
	if uid, ok := raw["uid"].(float64); ok {
		info.UserID = int64(uid)
	}
	if r, ok := raw["ratio"].(float64); ok {
		info.Ratio = r
	}
	if up, ok := raw["uploaded"].(string); ok {
		info.Uploaded = up
	}
	if down, ok := raw["downloaded"].(string); ok {
		info.Downloaded = down
	}
	if bp, ok := raw["seedbonus"].(float64); ok {
		info.BonusPoints = bp
	}
	if vip, ok := raw["vip"].(bool); ok {
		info.IsVIP = vip
	}

	return info, nil
}

func (c *Client) Search(ctx context.Context, query string, limit int) ([]TorrentItem, error) {
	if !c.HasAuth() {
		return nil, fmt.Errorf("mam_id não configurado")
	}

	if limit <= 0 {
		limit = 25
	}

	params := url.Values{}
	params.Set("tor[text]", query)
	params.Set("tor[browse_lang][]", "1") // Inglês e outros idiomas
	params.Set("tor[srchIn][title]", "true")
	params.Set("tor[srchIn][author]", "true")
	params.Set("perpage", fmt.Sprintf("%d", limit))

	reqURL := fmt.Sprintf("%s/json/load.php?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Calibre-Web-Automated/2026.5")
	req.Header.Set("Cookie", fmt.Sprintf("mam_id=%s", c.mamID))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na pesquisa MAM: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Data []struct {
			ID          int64  `json:"id"`
			Title       string `json:"title"`
			AuthorInfo  string `json:"author_info"`
			Category    string `json:"category"`
			Filetype    string `json:"filetype"`
			Size        string `json:"size"`
			Seeders     int    `json:"seeders"`
			Leechers    int    `json:"leechers"`
			VIP         int    `json:"vip"`
			Free        int    `json:"free"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("erro ao decodificar catálogo MAM: %w", err)
	}

	var items []TorrentItem
	for _, d := range data.Data {
		items = append(items, TorrentItem{
			ID:        d.ID,
			Title:     d.Title,
			Author:    d.AuthorInfo,
			FileType:  d.Filetype,
			Size:      d.Size,
			Seeders:   d.Seeders,
			Leechers:  d.Leechers,
			VIP:       d.VIP > 0,
			Freeleech: d.Free > 0,
		})
	}

	return items, nil
}

func (c *Client) DownloadTorrentToIngest(ctx context.Context, torrentID int64, ingestDir string) (string, error) {
	if !c.HasAuth() {
		return "", fmt.Errorf("mam_id não configurado")
	}

	reqURL := fmt.Sprintf("%s/tor/download.php?tid=%d", c.baseURL, torrentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Calibre-Web-Automated/2026.5")
	req.Header.Set("Cookie", fmt.Sprintf("mam_id=%s", c.mamID))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro no download do torrent MAM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download MAM falhou com status: %d", resp.StatusCode)
	}

	targetPath := filepath.Join(ingestDir, fmt.Sprintf("mam_%d.torrent", torrentID))
	f, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", err
	}

	return targetPath, nil
}

