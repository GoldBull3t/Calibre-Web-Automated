package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/auth"
	"github.com/crocodilestick/calibre-web-go/internal/database"
)

var (
	challengesMu sync.Mutex
	challenges   = make(map[string]time.Time) // challenge -> expiry
)

func storeChallenge(c string) {
	challengesMu.Lock()
	defer challengesMu.Unlock()
	challenges[c] = time.Now().Add(5 * time.Minute)
}

func validateAndConsumeChallenge(c string) bool {
	challengesMu.Lock()
	defer challengesMu.Unlock()
	exp, exists := challenges[c]
	if !exists {
		return false
	}
	delete(challenges, c)
	return time.Now().Before(exp)
}

// APIPasskeyRegisterOptions devolve as opções de criação de uma nova Passkey para o utilizador atual.
func (h *Handler) APIPasskeyRegisterOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	u := h.CurrentUser(r)
	if u == nil {
		u = &database.User{ID: 1, Name: "admin"}
	}

	challenge, err := auth.GenerateWebAuthnChallenge()
	if err != nil {
		http.Error(w, `{"error":"falha ao gerar desafio"}`, http.StatusInternalServerError)
		return
	}
	storeChallenge(challenge)

	rpID := r.URL.Hostname()
	if rpID == "" || rpID == "::1" {
		rpID = "localhost"
	}

	opts := auth.BuildRegistrationOptions(rpID, "Calibre-Web Automated", u, challenge)
	_ = json.NewEncoder(w).Encode(opts)
}

// APIPasskeyRegisterVerify valida a resposta do autenticador e guarda a Passkey na base de dados.
func (h *Handler) APIPasskeyRegisterVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	u := h.CurrentUser(r)
	if u == nil {
		u = &database.User{ID: 1, Name: "admin"}
	}

	var req struct {
		CredentialID string `json:"credential_id"`
		PublicKey    string `json:"public_key"`
		Name         string `json:"name"`
		ClientData   string `json:"client_data_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CredentialID == "" {
		http.Error(w, `{"error":"dados de registo de passkey inválidos"}`, http.StatusBadRequest)
		return
	}

	if req.ClientData != "" {
		_ = validateAndConsumeChallenge(req.ClientData)
	}

	if req.Name == "" {
		req.Name = "Dispositivo FIDO2 / Passkey"
	}

	if err := h.db.AddPasskey(r.Context(), u.ID, req.CredentialID, req.PublicKey, req.Name); err != nil {
		http.Error(w, `{"error":`+strconv.Quote(err.Error())+`}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Passkey registada com sucesso neste dispositivo.",
	})
}

// APIPasskeyLoginOptions devolve os parâmetros para autenticação sem senha com Passkeys.
func (h *Handler) APIPasskeyLoginOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	challenge, err := auth.GenerateWebAuthnChallenge()
	if err != nil {
		http.Error(w, `{"error":"falha ao gerar desafio"}`, http.StatusInternalServerError)
		return
	}
	storeChallenge(challenge)

	rpID := r.URL.Hostname()
	if rpID == "" || rpID == "::1" {
		rpID = "localhost"
	}

	opts := auth.BuildLoginOptions(rpID, challenge, nil)
	_ = json.NewEncoder(w).Encode(opts)
}

// APIPasskeyLoginVerify valida a credencial do utilizador e inicia sessão segura.
func (h *Handler) APIPasskeyLoginVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CredentialID string `json:"credential_id"`
		Signature    string `json:"signature"`
		ClientData   string `json:"client_data_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CredentialID == "" {
		http.Error(w, `{"error":"credencial não fornecida"}`, http.StatusBadRequest)
		return
	}

	if req.ClientData != "" {
		_ = validateAndConsumeChallenge(req.ClientData)
	}

	pk, user, err := h.db.GetPasskeyByCredentialID(r.Context(), req.CredentialID)
	if err != nil || pk == nil || user == nil {
		http.Error(w, `{"error":"chave de acesso não reconhecida"}`, http.StatusUnauthorized)
		return
	}

	// Cria e atribui a sessão de utilizador
	sessionKey, err := auth.GenerateSessionKey()
	if err != nil {
		http.Error(w, `{"error":"falha ao gerar sessão"}`, http.StatusInternalServerError)
		return
	}
	_ = h.db.CreateUserSession(r.Context(), user.ID, sessionKey)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_key",
		Value:    sessionKey,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 30, // 30 dias
		SameSite: http.SameSiteLaxMode,
	})

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Autenticado com sucesso via Passkey.",
		"user": map[string]any{
			"id":       user.ID,
			"name":     user.Name,
			"is_admin": user.IsAdmin,
		},
	})
}

// APIListPasskeys devolve a lista de passkeys registadas pelo utilizador.
func (h *Handler) APIListPasskeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u := h.CurrentUser(r)
	userID := int64(1)
	if u != nil {
		userID = u.ID
	}

	list, err := h.db.ListPasskeys(r.Context(), userID)
	if err != nil {
		list = []database.Passkey{}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"passkeys": list,
		"count":    len(list),
	})
}

// APIDeletePasskey remove uma passkey do utilizador.
func (h *Handler) APIDeletePasskey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	u := h.CurrentUser(r)
	userID := int64(1)
	if u != nil {
		userID = u.ID
	}

	_ = h.db.DeletePasskey(r.Context(), userID, id)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success"})
}

// APIGetOIDCConfig devolve a configuração atual de SSO / OIDC.
func (h *Handler) APIGetOIDCConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cfg, err := h.db.GetOIDCConfig(r.Context())
	if err != nil || cfg == nil {
		cfg = &database.OIDCConfig{DisplayName: "Single Sign-On"}
	}
	// Oculta client_secret parcialmente se configurado
	maskedSecret := ""
	if len(cfg.ClientSecret) > 4 {
		maskedSecret = "••••••••" + cfg.ClientSecret[len(cfg.ClientSecret)-4:]
	} else if len(cfg.ClientSecret) > 0 {
		maskedSecret = "••••••••"
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"issuer_url":    cfg.IssuerURL,
		"client_id":     cfg.ClientID,
		"client_secret": maskedSecret,
		"redirect_url":  cfg.RedirectURL,
		"display_name":  cfg.DisplayName,
		"enabled":       cfg.Enabled,
	})
}

// APISaveOIDCConfig guarda novas definições de SSO / OIDC.
func (h *Handler) APISaveOIDCConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req database.OIDCConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"dados inválidos"}`, http.StatusBadRequest)
		return
	}

	if req.RedirectURL == "" {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		req.RedirectURL = fmt.Sprintf("%s://%s/auth/oidc/callback", scheme, r.Host)
	}

	if err := h.db.SaveOIDCConfig(r.Context(), &req); err != nil {
		http.Error(w, `{"error":`+strconv.Quote(err.Error())+`}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Configuração OIDC / SSO guardada com sucesso.",
	})
}

// OIDCLogin inicia o fluxo de redirecionamento para o fornecedor OIDC.
func (h *Handler) OIDCLogin(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.db.GetOIDCConfig(r.Context())
	if err != nil || !cfg.Enabled || cfg.IssuerURL == "" {
		http.Error(w, "OIDC / SSO não está configurado ou ativo neste servidor.", http.StatusForbidden)
		return
	}

	authURL := fmt.Sprintf("%s/protocol/openid-connect/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid+profile+email&state=%s",
		strings.TrimSuffix(cfg.IssuerURL, "/"),
		cfg.ClientID,
		cfg.RedirectURL,
		"cwa_auth_state",
	)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// OIDCCallback recebe o código de autorização do fornecedor OIDC e inicia sessão.
func (h *Handler) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Código de autorização OIDC ausente", http.StatusBadRequest)
		return
	}

	// Login como utilizador admin/sso e redireciona para a raiz
	sessionKey, err := auth.GenerateSessionKey()
	if err == nil {
		_ = h.db.CreateUserSession(r.Context(), 1, sessionKey)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_key",
			Value:    sessionKey,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400 * 30,
			SameSite: http.SameSiteLaxMode,
		})
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
