package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/auth"
)

type OIDCConfig struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
}

type OAuthProviderConfig struct {
	ProviderName string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
}

var (
	oidcCache   = make(map[string]*OIDCConfig)
	oidcCacheMu sync.RWMutex
)

// DiscoverOIDC obtém as URLs dos endpoints através do padrão .well-known/openid-configuration.
func DiscoverOIDC(ctx context.Context, issuerURL string) (*OIDCConfig, error) {
	issuerURL = strings.TrimSuffix(issuerURL, "/")
	wellKnown := issuerURL + "/.well-known/openid-configuration"

	oidcCacheMu.RLock()
	if cfg, ok := oidcCache[wellKnown]; ok {
		oidcCacheMu.RUnlock()
		return cfg, nil
	}
	oidcCacheMu.RUnlock()

	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", wellKnown, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC discovery retornou status HTTP %d", resp.StatusCode)
	}

	var cfg OIDCConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, err
	}

	oidcCacheMu.Lock()
	oidcCache[wellKnown] = &cfg
	oidcCacheMu.Unlock()

	return &cfg, nil
}

// OAuthLogin inicia o fluxo de autenticação OAuth 2.0 / OIDC redirecionando o utilizador para o IdP.
func (h *Handler) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	if provider == "" {
		provider = r.URL.Query().Get("provider")
	}
	if provider == "" {
		provider = "google"
	}

	stateBytes := make([]byte, 16)
	_, _ = rand.Read(stateBytes)
	state := hex.EncodeToString(stateBytes)

	// Guarda state num cookie temporário seguro
	http.SetCookie(w, &http.Cookie{
		Name:     "cwa_oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   300,
		SameSite: http.SameSiteLaxMode,
	})

	cfg := h.getProviderConfig(r.Context(), provider, r)
	if cfg == nil || cfg.AuthURL == "" {
		http.Error(w, fmt.Sprintf("Provedor OAuth %q não configurado", provider), http.StatusBadRequest)
		return
	}

	redirectURI := h.getOAuthRedirectURI(r, provider)
	scope := strings.Join(cfg.Scopes, " ")
	if scope == "" {
		scope = "openid profile email"
	}

	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		cfg.AuthURL,
		url.QueryEscape(cfg.ClientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(scope),
		url.QueryEscape(state),
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

// OAuthCallback processa o retorno do IdP com o código de autorização, valida e inicia sessão.
func (h *Handler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	if provider == "" {
		provider = r.URL.Query().Get("provider")
	}
	if provider == "" {
		provider = "google"
	}

	// 1. Validação de CSRF state
	stateQuery := r.URL.Query().Get("state")
	stateCookie, err := r.Cookie("cwa_oauth_state")
	if err != nil || stateCookie.Value == "" || stateCookie.Value != stateQuery {
		http.Error(w, "Falha de validação OAuth: estado CSRF inválido ou expirado.", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		errDesc := r.URL.Query().Get("error_description")
		http.Error(w, fmt.Sprintf("Autenticação OAuth cancelada pelo provedor: %s", errDesc), http.StatusBadRequest)
		return
	}

	cfg := h.getProviderConfig(r.Context(), provider, r)
	if cfg == nil || cfg.TokenURL == "" {
		http.Error(w, "Configuração do provedor ausente", http.StatusInternalServerError)
		return
	}

	redirectURI := h.getOAuthRedirectURI(r, provider)

	// 2. Troca code por token
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
	}

	client := &http.Client{Timeout: 10 * time.Second}
	tokenReq, err := http.NewRequestWithContext(r.Context(), "POST", cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		http.Error(w, "Erro ao preparar requisição de token", http.StatusInternalServerError)
		return
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("Accept", "application/json")

	tokenResp, err := client.Do(tokenReq)
	if err != nil || tokenResp.StatusCode != http.StatusOK {
		http.Error(w, "Falha ao obter token de acesso do provedor OAuth", http.StatusBadGateway)
		return
	}
	defer tokenResp.Body.Close()

	var tokenResult struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResult); err != nil || tokenResult.AccessToken == "" {
		http.Error(w, "Resposta de token inválida", http.StatusBadGateway)
		return
	}

	// 3. Obtém dados do utilizador (userinfo)
	userInfoReq, err := http.NewRequestWithContext(r.Context(), "GET", cfg.UserInfoURL, nil)
	if err != nil {
		http.Error(w, "Falha ao preparar requisição de perfil", http.StatusInternalServerError)
		return
	}
	userInfoReq.Header.Set("Authorization", "Bearer "+tokenResult.AccessToken)
	userInfoReq.Header.Set("Accept", "application/json")

	userInfoResp, err := client.Do(userInfoReq)
	if err != nil || userInfoResp.StatusCode != http.StatusOK {
		http.Error(w, "Falha ao carregar informações de perfil do utilizador", http.StatusBadGateway)
		return
	}
	defer userInfoResp.Body.Close()

	var userInfo struct {
		Sub               string `json:"sub"`
		Email             string `json:"email"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
		Login             string `json:"login"` // GitHub usa 'login'
	}
	_ = json.NewDecoder(userInfoResp.Body).Decode(&userInfo)

	username := userInfo.PreferredUsername
	if username == "" {
		username = userInfo.Login
	}
	if username == "" {
		username = userInfo.Name
	}
	if username == "" && userInfo.Email != "" {
		username = strings.Split(userInfo.Email, "@")[0]
	}
	if username == "" {
		username = "oauth_" + userInfo.Sub
	}

	// 4. Criação ou recuperação de utilizador na base de dados
	user, err := h.db.GetUserByName(r.Context(), username)
	if err != nil || user == nil {
		if userInfo.Email != "" {
			user, _ = h.db.GetUserByEmail(r.Context(), userInfo.Email)
		}
	}

	if user == nil {
		// Registo automático de novo utilizador OAuth
		pwdHash, _ := auth.HashPassword("oauth_ext_" + stateQuery)
		newUserID, err := h.db.CreateUser(r.Context(), username, userInfo.Email, pwdHash, auth.RoleViewer|auth.RoleDownload)
		if err != nil {
			http.Error(w, "Falha ao aprovisionar utilizador OAuth", http.StatusInternalServerError)
			return
		}
		user, _ = h.db.GetUserByID(r.Context(), newUserID)
	}

	if user == nil {
		http.Error(w, "Utilizador não encontrado", http.StatusUnauthorized)
		return
	}

	// 5. Inicia sessão do utilizador com cookie oficial
	sessionKey, err := auth.GenerateSessionKey()
	if err == nil {
		_ = h.db.CreateUserSession(r.Context(), user.ID, sessionKey)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionKey,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 30, // 30 dias de validade
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) getOAuthRedirectURI(r *http.Request, provider string) string {
	// Se especificado OAUTH_REDIRECT_HOST pelo utilizador/reverse-proxy
	if host := os.Getenv("OAUTH_REDIRECT_HOST"); host != "" {
		return fmt.Sprintf("%s/oauth/%s/callback", strings.TrimSuffix(host, "/"), provider)
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/oauth/%s/callback", scheme, r.Host, provider)
}

func (h *Handler) getProviderConfig(ctx context.Context, provider string, r *http.Request) *OAuthProviderConfig {
	provider = strings.ToLower(provider)

	// Suporte a Generic OIDC Discovery (Keycloak, Authentik, etc.)
	if oidcIssuer := os.Getenv("OIDC_ISSUER_URL"); oidcIssuer != "" && (provider == "oidc" || provider == "generic") {
		if disco, err := DiscoverOIDC(ctx, oidcIssuer); err == nil {
			return &OAuthProviderConfig{
				ProviderName: "oidc",
				ClientID:     os.Getenv("OIDC_CLIENT_ID"),
				ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
				AuthURL:      disco.AuthorizationEndpoint,
				TokenURL:     disco.TokenEndpoint,
				UserInfoURL:  disco.UserinfoEndpoint,
				Scopes:       []string{"openid", "profile", "email"},
			}
		}
	}

	switch provider {
	case "google":
		return &OAuthProviderConfig{
			ProviderName: "google",
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			UserInfoURL:  "https://openidconnect.googleapis.com/v1/userinfo",
			Scopes:       []string{"openid", "email", "profile"},
		}
	case "github":
		return &OAuthProviderConfig{
			ProviderName: "github",
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
			UserInfoURL:  "https://api.github.com/user",
			Scopes:       []string{"read:user", "user:email"},
		}
	default:
		return nil
	}
}
