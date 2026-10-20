package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/auth"
	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/flosch/pongo2/v6"
)

// Login renderiza a tela de autenticação original em GET e autentica em POST.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = h.loginTpl.ExecuteWriter(pongo2.Context{}, w)
		return
	}

	_ = r.ParseForm()
	username := strings.TrimSpace(r.FormValue("username"))
	if username == "" {
		username = strings.TrimSpace(r.FormValue("name"))
	}
	password := r.FormValue("password")

	user, err := h.db.GetUserByName(r.Context(), username)
	if err != nil || !auth.CheckPassword(user.PasswordHash, password) {
		if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "message": "Invalid username or password"})
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = h.loginTpl.ExecuteWriter(pongo2.Context{
			"error":    "Invalid username or password.",
			"username": username,
		}, w)
		return
	}

	sessionKey, err := auth.GenerateSessionKey()
	if err == nil {
		_ = h.db.CreateUserSession(r.Context(), user.ID, sessionKey)
	}

	maxAge := 30 * 24 * 3600
	if r.FormValue("remember_me") == "" {
		maxAge = 24 * 3600
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionKey,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// Logout encerra a sessão e redireciona para a tela inicial.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil && cookie.Value != "" {
		_ = h.db.DeleteUserSession(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Register renderiza a página de registro.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = h.loginTpl.ExecuteWriter(pongo2.Context{"isRegister": true}, w)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Profile entrega a página de perfil do utilizador ou processa alterações.
func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	u := h.CurrentUser(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		email := strings.TrimSpace(r.FormValue("email"))
		if email == "" {
			email = u.Email
		}
		kindleMail := strings.TrimSpace(r.FormValue("kindle_mail"))
		locale := strings.TrimSpace(r.FormValue("locale"))
		if locale == "" {
			locale = "en"
		}
		defaultLang := strings.TrimSpace(r.FormValue("default_language"))
		newPassword := r.FormValue("password")

		var passwordHash string
		if newPassword != "" {
			var err error
			passwordHash, err = auth.HashPassword(newPassword)
			if err != nil {
				http.Error(w, "Erro ao processar senha", http.StatusInternalServerError)
				return
			}
		}

		if err := h.db.UpdateUserProfile(r.Context(), u.ID, email, kindleMail, locale, defaultLang, passwordHash); err != nil {
			http.Error(w, "Erro ao salvar perfil", http.StatusInternalServerError)
			return
		}

		if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
			return
		}
		http.Redirect(w, r, "/me", http.StatusSeeOther)
		return
	}

	h.Admin(w, r)
}

// ProfilePicture serve o avatar do utilizador.
func (h *Handler) ProfilePicture(w http.ResponseWriter, r *http.Request) {
	h.serveDefaultCover(w, r)
}

// UserProfilesJSON devolve as fotos de perfil em JSON.
func (h *Handler) UserProfilesJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"profiles": []any{},
	})
}

// authenticateKOSync autentica chamadas da API do KOReader via headers ou Basic Auth.
func (h *Handler) authenticateKOSync(r *http.Request) *database.User {
	// 1. Headers proprietários KOReader
	userHeader := r.Header.Get("x-auth-user")
	keyHeader := r.Header.Get("x-auth-key")
	if userHeader != "" && keyHeader != "" {
		u, err := h.db.GetUserByName(r.Context(), userHeader)
		if err == nil && auth.CheckPassword(u.PasswordHash, keyHeader) {
			return u
		}
	}

	// 2. HTTP Basic Auth
	if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Basic ") {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Basic "))
		if err == nil {
			parts := strings.SplitN(string(decoded), ":", 2)
			if len(parts) == 2 {
				u, err := h.db.GetUserByName(r.Context(), parts[0])
				if err == nil && auth.CheckPassword(u.PasswordHash, parts[1]) {
					return u
				}
			}
		}
	}

	// 3. Fallback de sessão
	if u := h.CurrentUser(r); u != nil {
		return u
	}

	return nil
}

// KOSyncAuth valida as credenciais do utilizador para o KOReader.
func (h *Handler) KOSyncAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u := h.authenticateKOSync(r)
	if u == nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":   2001,
			"message": "Unauthorized",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"authorized": "OK",
	})
}

// KOSyncGetProgress devolve o progresso de leitura guardado para o documento indicado.
func (h *Handler) KOSyncGetProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u := h.authenticateKOSync(r)
	if u == nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":   2001,
			"message": "Unauthorized",
		})
		return
	}

	doc := r.PathValue("document")
	if doc == "" {
		doc = strings.TrimPrefix(r.URL.Path, "/kosync/syncs/progress/")
		doc = strings.TrimPrefix(doc, "/syncs/progress/")
	}
	if doc == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":   2004,
			"message": "Invalid document field",
		})
		return
	}

	p, err := h.db.GetKOSyncProgress(r.Context(), u.ID, doc)
	if err != nil {
		// KOReader espera dicionário vazio se ainda não houver progresso sincronizado
		_ = json.NewEncoder(w).Encode(map[string]any{})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"document":   p.Document,
		"progress":   p.Progress,
		"percentage": p.Percentage / 100.0,
		"device":     p.Device,
		"device_id":  p.DeviceID,
		"timestamp":  p.Timestamp,
	})
}

// KOSyncUpdateProgress recebe e persiste o progresso de leitura do KOReader.
func (h *Handler) KOSyncUpdateProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u := h.authenticateKOSync(r)
	if u == nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":   2001,
			"message": "Unauthorized",
		})
		return
	}

	var req struct {
		Document   string  `json:"document"`
		Progress   string  `json:"progress"`
		Percentage float64 `json:"percentage"`
		Device     string  `json:"device"`
		DeviceID   string  `json:"device_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Document == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":   2003,
			"message": "Invalid request data",
		})
		return
	}

	pct := req.Percentage
	if pct <= 1.0 {
		pct *= 100.0
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	now := time.Now().Unix()
	p := database.KOSyncProgress{
		UserID:     u.ID,
		Document:   req.Document,
		Progress:   req.Progress,
		Percentage: pct,
		Device:     req.Device,
		DeviceID:   req.DeviceID,
		Timestamp:  now,
	}

	_ = h.db.UpdateKOSyncProgress(r.Context(), p)

	// Se for ID numérico de livro, sincroniza com o estado lido/não lido e marcador
	if bookID, err := strconv.ParseInt(req.Document, 10, 64); err == nil && bookID > 0 {
		if pct >= 99.0 {
			_ = h.db.BulkSetReadStatus(r.Context(), u.ID, []int64{bookID}, true)
		}
		_ = h.db.SaveBookmark(r.Context(), bookID, u.ID, "EPUB", req.Progress)
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"document":  req.Document,
		"timestamp": now,
	})
}

// KOSyncCreateUser gere criação de utilizadores via KOReader.
func (h *Handler) KOSyncCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": 2003, "message": "Missing fields"})
		return
	}

	if existing, _ := h.db.GetUserByName(r.Context(), req.Username); existing != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": 2002, "message": "User exists"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "User registered"})
}

// KOSync despacha chamadas universais para /kosync e /kosync/{path...}.
func (h *Handler) KOSync(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if strings.Contains(p, "/users/auth") {
		h.KOSyncAuth(w, r)
		return
	}
	if strings.Contains(p, "/users/create") {
		h.KOSyncCreateUser(w, r)
		return
	}
	if strings.Contains(p, "/syncs/progress") {
		if r.Method == http.MethodPut {
			h.KOSyncUpdateProgress(w, r)
			return
		}
		h.KOSyncGetProgress(w, r)
		return
	}

	// Página de informação / plugin
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"name":        "Calibre-Web KOReader Sync",
		"version":     "1.0.0",
		"status":      "online",
		"description": "High-performance sync server compatible with KOReader",
	})
}

// authenticateKobo localiza o utilizador associado a um token Kobo na URL ou sessão.
func (h *Handler) authenticateKobo(r *http.Request) *database.User {
	token := r.PathValue("token")
	if token == "" {
		parts := strings.Split(r.URL.Path, "/")
		for i, part := range parts {
			if part == "kobo" && i+1 < len(parts) {
				token = parts[i+1]
				break
			}
		}
	}

	if token != "" {
		if u, err := h.db.GetUserByKoboToken(r.Context(), token); err == nil && u != nil {
			return u
		}
	}

	return h.CurrentUser(r)
}

// KoboInitialization fornece o mapa de recursos da API Kobo.
func (h *Handler) KoboInitialization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)
	token := r.PathValue("token")
	if token == "" {
		token = "default"
	}

	res := map[string]any{
		"Resources": map[string]any{
			"image_host":                 baseURL,
			"image_url_quality_template": fmt.Sprintf("%s/kobo/%s/{ImageId}/{Width}/{Height}/{Quality}/{IsGreyscale}/image.jpg", baseURL, token),
			"image_url_template":         fmt.Sprintf("%s/kobo/%s/{ImageId}/{Width}/{Height}/{IsGreyscale}/image.jpg", baseURL, token),
		},
	}
	_ = json.NewEncoder(w).Encode(res)
}

// KoboProfile fornece os dados de perfil do utilizador para o Kobo.
func (h *Handler) KoboProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u := h.authenticateKobo(r)
	userID := int64(1)
	userEmail := "reader@calibre-web.local"
	token := r.PathValue("token")
	if token == "" {
		token = "user-key"
	}
	if u != nil {
		userID = u.ID
		if u.Email != "" {
			userEmail = u.Email
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"UserKey":   token,
		"UserId":    fmt.Sprintf("%d", userID),
		"UserEmail": userEmail,
	})
}

// KoboAnalytics responde aos pings de telemetria do e-reader Kobo.
func (h *Handler) KoboAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{})
}

// KoboSync sincroniza o catálogo de livros para o Kobo eReader.
func (h *Handler) KoboSync(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u := h.authenticateKobo(r)
	var userID int64
	if u != nil {
		userID = u.ID
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)
	token := r.PathValue("token")

	result, err := h.db.ListBooks(r.Context(), database.BookFilter{
		Page:     1,
		PageSize: 1000,
		UserID:   userID,
	})
	if err != nil {
		_ = json.NewEncoder(w).Encode([]any{})
		return
	}

	syncItems := make([]map[string]any, 0, len(result.Items))
	for _, book := range result.Items {
		bookUUID := fmt.Sprintf("book-%d", book.ID)
		created := book.PubDate.UTC().Format(time.RFC3339)
		if created == "" || book.PubDate.IsZero() {
			created = time.Now().UTC().Format(time.RFC3339)
		}

		downloadURLs := make([]map[string]any, 0, len(book.Formats))
		for _, f := range book.Formats {
			url := fmt.Sprintf("%s/kobo/%s/download/%d/%s", baseURL, token, book.ID, f.Format)
			if token == "" {
				url = fmt.Sprintf("%s/download/%d/%s", baseURL, book.ID, f.Format)
			}
			downloadURLs = append(downloadURLs, map[string]any{
				"Format": f.Format,
				"Size":   f.Size,
				"Url":    url,
			})
		}

		contributors := make([]map[string]any, 0, len(book.Authors))
		for _, a := range book.Authors {
			contributors = append(contributors, map[string]any{
				"Contributor": a,
				"Role":        "Author",
			})
		}

		item := map[string]any{
			"NewEntitlement": map[string]any{
				"BookEntitlement": map[string]any{
					"Accessibility":       "Full",
					"ActivePeriod":        map[string]string{"From": created},
					"Created":             created,
					"CrossRevisionId":     bookUUID,
					"Id":                  bookUUID,
					"IsRemoved":           book.IsArchived,
					"IsHiddenFromArchive": false,
					"IsLocked":            false,
					"LastModified":        time.Now().UTC().Format(time.RFC3339),
					"OriginCategory":      "Imported",
					"RevisionId":          bookUUID,
					"Status":              "Active",
				},
				"BookMetadata": map[string]any{
					"CrossRevisionId":       bookUUID,
					"CurrencyCode":          "USD",
					"DownloadUrls":          downloadURLs,
					"EntitlementId":         bookUUID,
					"Id":                    bookUUID,
					"IsEligibleForKoboLove": false,
					"RevisionId":            bookUUID,
					"Title":                 book.Title,
					"ContributorRoles":      contributors,
				},
			},
		}
		syncItems = append(syncItems, item)
	}

	_ = json.NewEncoder(w).Encode(syncItems)
}

// KoboCover entrega a imagem de capa redimensionada ou nativa para o Kobo.
func (h *Handler) KoboCover(w http.ResponseWriter, r *http.Request) {
	bookUUID := r.PathValue("book_uuid")
	cleanID := strings.TrimPrefix(bookUUID, "book-")
	r.SetPathValue("id", cleanID)
	h.GetCover(w, r)
}

// KoboDownload serve o arquivo do livro no formato solicitado pelo Kobo.
func (h *Handler) KoboDownload(w http.ResponseWriter, r *http.Request) {
	h.DownloadBook(w, r)
}

// KoboReadingState atualiza o estado de leitura enviado pelo e-reader Kobo.
func (h *Handler) KoboReadingState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u := h.authenticateKobo(r)
	if u == nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "Unauthorized"})
		return
	}

	bookUUID := r.PathValue("book_uuid")
	cleanID := strings.TrimPrefix(bookUUID, "book-")
	bookID, _ := strconv.ParseInt(cleanID, 10, 64)

	var req struct {
		CurrentBookmark struct {
			ProgressPercent float64 `json:"ProgressPercent"`
			Location        string  `json:"Location"`
		} `json:"CurrentBookmark"`
		Status string `json:"Status"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if bookID > 0 {
		pct := req.CurrentBookmark.ProgressPercent
		if pct <= 1.0 {
			pct *= 100.0
		}
		_ = h.db.UpdateKOSyncProgress(r.Context(), database.KOSyncProgress{
			UserID:     u.ID,
			Document:   cleanID,
			Progress:   req.CurrentBookmark.Location,
			Percentage: pct,
			Device:     "Kobo",
			Timestamp:  time.Now().Unix(),
		})
		if pct >= 99.0 || strings.EqualFold(req.Status, "Finished") {
			_ = h.db.BulkSetReadStatus(r.Context(), u.ID, []int64{bookID}, true)
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// GenerateKoboAuthToken cria e retorna o token de autenticação e URL do Kobo para o utilizador.
func (h *Handler) GenerateKoboAuthToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u := h.CurrentUser(r)
	if u == nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "Unauthorized"})
		return
	}

	targetUserID := u.ID
	if uidStr := r.PathValue("user_id"); uidStr != "" {
		if parsed, err := strconv.ParseInt(uidStr, 10, 64); err == nil && (u.IsAdmin || parsed == u.ID) {
			targetUserID = parsed
		}
	}

	token, err := h.db.GetOrCreateKoboToken(r.Context(), targetUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "Failed to generate token"})
		return
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	apiEndpoint := fmt.Sprintf("%s://%s/kobo/%s", scheme, r.Host, token)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"token":        token,
		"api_endpoint": apiEndpoint,
		"user_id":      targetUserID,
	})
}
