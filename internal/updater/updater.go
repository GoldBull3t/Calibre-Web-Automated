package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// CurrentVersion define a versão atual do Calibre-Web Automated Go.
var CurrentVersion = "v2.3.0"

const (
	GithubReleaseURL = "https://api.github.com/repos/crocodilestick/calibre-web-automated/releases/latest"
	cacheDuration    = 24 * time.Hour
)

type UpdateInfo struct {
	UpdateAvailable bool      `json:"update_available"`
	CurrentVersion  string    `json:"current_version"`
	LatestVersion   string    `json:"latest_version"`
	ReleaseURL      string    `json:"release_url"`
	ReleaseNotes    string    `json:"release_notes"`
	CheckedAt       time.Time `json:"checked_at"`
}

type Updater struct {
	client     *http.Client
	mu         sync.RWMutex
	cachedInfo *UpdateInfo
}

var defaultUpdater = &Updater{
	client: &http.Client{Timeout: 5 * time.Second},
}

func GetDefault() *Updater {
	return defaultUpdater
}

type githubReleaseResponse struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

// CheckUpdate consulta a API do GitHub apenas uma vez a cada 24 horas (em cache).
func (u *Updater) CheckUpdate(ctx context.Context) (*UpdateInfo, error) {
	u.mu.RLock()
	if u.cachedInfo != nil && time.Since(u.cachedInfo.CheckedAt) < cacheDuration {
		info := *u.cachedInfo
		u.mu.RUnlock()
		return &info, nil
	}
	u.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "GET", GithubReleaseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Calibre-Web-Automated-Go/"+CurrentVersion)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := u.client.Do(req)
	if err != nil {
		// Se offline ou falha de rede, devolve o cache anterior se existir ou resposta neutra
		u.mu.RLock()
		if u.cachedInfo != nil {
			info := *u.cachedInfo
			u.mu.RUnlock()
			return &info, nil
		}
		u.mu.RUnlock()
		return &UpdateInfo{
			CurrentVersion: CurrentVersion,
			LatestVersion:  CurrentVersion,
			CheckedAt:      time.Now(),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &UpdateInfo{
			CurrentVersion: CurrentVersion,
			LatestVersion:  CurrentVersion,
			CheckedAt:      time.Now(),
		}, nil
	}

	var ghResp githubReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghResp); err != nil {
		return nil, fmt.Errorf("falha ao interpretar JSON de release: %w", err)
	}

	latestTag := strings.TrimSpace(ghResp.TagName)
	updateAvailable := isNewerVersion(CurrentVersion, latestTag)

	info := &UpdateInfo{
		UpdateAvailable: updateAvailable,
		CurrentVersion:  CurrentVersion,
		LatestVersion:   latestTag,
		ReleaseURL:      ghResp.HTMLURL,
		ReleaseNotes:    ghResp.Body,
		CheckedAt:       time.Now(),
	}

	u.mu.Lock()
	u.cachedInfo = info
	u.mu.Unlock()

	return info, nil
}

func isNewerVersion(current, latest string) bool {
	c := strings.TrimPrefix(strings.TrimSpace(current), "v")
	l := strings.TrimPrefix(strings.TrimSpace(latest), "v")
	if c == "" || l == "" {
		return false
	}
	return c != l && l > c
}

