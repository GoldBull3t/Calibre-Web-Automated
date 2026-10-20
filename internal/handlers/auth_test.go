package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/crocodilestick/calibre-web-go/internal/handlers"
)

func setupTestDB(t *testing.T) (*database.DB, *handlers.Handler, func()) {
	tmpDir, err := os.MkdirTemp("", "calibre-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	metadataSrc := filepath.Join("..", "..", "empty_library", "metadata.db")
	appSrc := filepath.Join("..", "..", "empty_library", "app.db")

	metadataDst := filepath.Join(tmpDir, "metadata.db")
	appDst := filepath.Join(tmpDir, "app.db")

	metaBytes, err := os.ReadFile(metadataSrc)
	if err != nil {
		t.Fatalf("Failed to read metadata.db: %v", err)
	}
	if err := os.WriteFile(metadataDst, metaBytes, 0644); err != nil {
		t.Fatalf("Failed to copy metadata.db: %v", err)
	}

	appBytes, err := os.ReadFile(appSrc)
	if err != nil {
		t.Fatalf("Failed to read app.db: %v", err)
	}
	if err := os.WriteFile(appDst, appBytes, 0644); err != nil {
		t.Fatalf("Failed to copy app.db: %v", err)
	}

	db, err := database.Open(metadataDst, appDst)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}

	h, err := handlers.New(db, tmpDir, tmpDir, filepath.Join("..", "..", "cps", "static"), filepath.Join(tmpDir, "ingest"), filepath.Join(tmpDir, "plugins"))
	if err != nil {
		t.Fatalf("Failed to create handlers: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, h, cleanup
}

func TestLoginAndSessionLifecycle(t *testing.T) {
	_, h, cleanup := setupTestDB(t)
	defer cleanup()

	// 1. Login com credenciais incorretas
	form := url.Values{
		"username": {"admin"},
		"password": {"wrongpassword"},
	}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Login(w, req)
	if w.Code == http.StatusFound {
		t.Fatal("Login com senha incorreta não deveria redirecionar para a home")
	}

	// 2. Login com credenciais corretas do Calibre-Web (admin / admin123)
	formValid := url.Values{
		"username": {"admin"},
		"password": {"admin123"},
	}
	reqValid := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(formValid.Encode()))
	reqValid.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	wValid := httptest.NewRecorder()

	h.Login(wValid, reqValid)
	if wValid.Code != http.StatusFound {
		t.Fatalf("Esperado redirect 302, recebido: %d", wValid.Code)
	}

	cookies := wValid.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatal("Cookie de sessão não foi emitido após login válido")
	}

	// 3. Teste do Middleware com a sessão criada
	mux := http.NewServeMux()
	mux.HandleFunc("/test-auth", func(w http.ResponseWriter, r *http.Request) {
		u := h.CurrentUser(r)
		if u == nil || u.Name != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	handlerWithAuth := h.AuthMiddleware(mux)
	testReq := httptest.NewRequest(http.MethodGet, "/test-auth", nil)
	testReq.AddCookie(sessionCookie)
	wTest := httptest.NewRecorder()

	handlerWithAuth.ServeHTTP(wTest, testReq)
	if wTest.Code != http.StatusOK {
		t.Fatalf("Esperado 200 OK via AuthMiddleware com sessão válida, recebido: %d", wTest.Code)
	}
}

func TestToggleReadAndBookmarks(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	// Autentica admin
	user, err := db.GetUserByName(context.Background(), "admin")
	if err != nil || user == nil {
		t.Fatalf("Admin não encontrado no DB de teste: %v", err)
	}

	// 1. Alterna leitura do livro 1
	status1, err := db.ToggleBookRead(context.Background(), user.ID, 1)
	if err != nil {
		t.Fatalf("Erro ao alternar leitura do livro: %v", err)
	}
	if !status1 {
		t.Fatal("Esperado que livro ficasse lido (true)")
	}

	// Verifica se GetBookUserState reflete o estado lido
	isRead, _, err := db.GetBookUserState(context.Background(), user.ID, 1)
	if err != nil || !isRead {
		t.Fatalf("GetBookUserState deveria retornar lido=true: %v, %v", isRead, err)
	}

	// Alterna novamente -> deve voltar a não-lido
	status2, err := db.ToggleBookRead(context.Background(), user.ID, 1)
	if err != nil {
		t.Fatalf("Erro ao alternar leitura: %v", err)
	}
	if status2 {
		t.Fatal("Esperado que livro voltasse a não lido (false)")
	}

	// 2. Salva bookmark
	cfi := "epubcfi(/6/4[chapter-1]!/4/2/10)"
	if err := db.SaveBookmark(context.Background(), user.ID, 1, "EPUB", cfi); err != nil {
		t.Fatalf("Erro ao salvar bookmark: %v", err)
	}

	retrieved, err := db.GetBookmark(context.Background(), user.ID, 1, "EPUB")
	if err != nil {
		t.Fatalf("Erro ao recuperar bookmark: %v", err)
	}
	if retrieved != cfi {
		t.Fatalf("Bookmark recuperado incorreto: esperado %s, obtido %s", cfi, retrieved)
	}
}

func TestBulkOperationsAndShelves(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	user, err := db.GetUserByName(context.Background(), "admin")
	if err != nil || user == nil {
		t.Fatalf("Admin não encontrado: %v", err)
	}

	// 1. Teste de operações em lote (Read & Archive)
	bookIDs := []int64{1, 2}
	if err := db.BulkSetReadStatus(context.Background(), user.ID, bookIDs, true); err != nil {
		t.Fatalf("Erro no BulkSetReadStatus: %v", err)
	}
	isRead1, _, _ := db.GetBookUserState(context.Background(), user.ID, 1)
	isRead2, _, _ := db.GetBookUserState(context.Background(), user.ID, 2)
	if !isRead1 || !isRead2 {
		t.Fatalf("Livros 1 e 2 deveriam estar marcados como lidos: %v, %v", isRead1, isRead2)
	}

	if err := db.BulkSetArchived(context.Background(), user.ID, bookIDs, true); err != nil {
		t.Fatalf("Erro no BulkSetArchived: %v", err)
	}
	_, isArchived1, _ := db.GetBookUserState(context.Background(), user.ID, 1)
	if !isArchived1 {
		t.Fatal("Livro 1 deveria estar arquivado")
	}

	// 2. Teste de Estante e Reordenação
	shelfID, err := db.CreateShelf(context.Background(), user.ID, "Minha Estante VIP", true)
	if err != nil || shelfID <= 0 {
		t.Fatalf("Falha ao criar estante: %v", err)
	}

	_ = db.AddBookToShelf(context.Background(), shelfID, 1)
	_ = db.AddBookToShelf(context.Background(), shelfID, 2)

	// Reordena
	orderMap := map[int64]int{1: 2, 2: 1}
	if err := db.ReorderShelfBooks(context.Background(), shelfID, orderMap); err != nil {
		t.Fatalf("Erro ao reordenar estante: %v", err)
	}
}

func TestKOSyncAndKoboSync(t *testing.T) {
	db, h, cleanup := setupTestDB(t)
	defer cleanup()

	// 1. KOReader Auth: credenciais inválidas
	reqBad := httptest.NewRequest(http.MethodGet, "/users/auth", nil)
	reqBad.Header.Set("x-auth-user", "admin")
	reqBad.Header.Set("x-auth-key", "wrong")
	wBad := httptest.NewRecorder()
	h.KOSyncAuth(wBad, reqBad)
	if wBad.Code != http.StatusUnauthorized {
		t.Fatalf("Esperado 401 para credenciais inválidas, recebido: %d", wBad.Code)
	}

	// 2. KOReader Auth: credenciais válidas
	reqAuth := httptest.NewRequest(http.MethodGet, "/users/auth", nil)
	reqAuth.Header.Set("x-auth-user", "admin")
	reqAuth.Header.Set("x-auth-key", "admin123")
	wAuth := httptest.NewRecorder()
	h.KOSyncAuth(wAuth, reqAuth)
	if wAuth.Code != http.StatusOK {
		t.Fatalf("Esperado 200 para credenciais válidas, recebido: %d", wAuth.Code)
	}

	// 3. KOReader Update Progress (99.5% de progresso marca o livro 1 como lido)
	body := `{"document":"1","progress":"/6/4/2[chapter1]","percentage":0.995,"device":"koreader-test"}`
	reqProg := httptest.NewRequest(http.MethodPut, "/syncs/progress", strings.NewReader(body))
	reqProg.Header.Set("Content-Type", "application/json")
	reqProg.Header.Set("x-auth-user", "admin")
	reqProg.Header.Set("x-auth-key", "admin123")
	wProg := httptest.NewRecorder()
	h.KOSyncUpdateProgress(wProg, reqProg)
	if wProg.Code != http.StatusOK {
		t.Fatalf("Falha no update de progresso KOSync: %d %s", wProg.Code, wProg.Body.String())
	}

	// Verifica se o livro 1 foi marcado como lido
	user, err := db.GetUserByName(context.Background(), "admin")
	if err != nil {
		t.Fatalf("Erro ao obter utilizador admin: %v", err)
	}
	isRead, _, err := db.GetBookUserState(context.Background(), user.ID, 1)
	if err != nil || !isRead {
		t.Fatalf("Livro 1 deveria ter sido marcado como lido após sync >= 99%%: isRead=%v, err=%v", isRead, err)
	}

	// 4. KOReader Get Progress
	reqGetProg := httptest.NewRequest(http.MethodGet, "/syncs/progress/1", nil)
	reqGetProg.SetPathValue("document", "1")
	reqGetProg.Header.Set("x-auth-user", "admin")
	reqGetProg.Header.Set("x-auth-key", "admin123")
	wGetProg := httptest.NewRecorder()
	h.KOSyncGetProgress(wGetProg, reqGetProg)
	if wGetProg.Code != http.StatusOK {
		t.Fatalf("Falha no get de progresso KOSync: %d", wGetProg.Code)
	}
	var progResp map[string]any
	if err := json.Unmarshal(wGetProg.Body.Bytes(), &progResp); err != nil {
		t.Fatalf("Resposta KOSync inválida: %v", err)
	}
	if progResp["document"] != "1" || progResp["device"] != "koreader-test" {
		t.Fatalf("Dados de progresso KOSync incorretos: %+v", progResp)
	}

	// 5. Kobo Auth: Gerar token para o utilizador admin
	reqTok := httptest.NewRequest(http.MethodGet, "/kobo_auth/generate_auth_token/1", nil)
	// Simula login de sessão admin
	sessionKey := "test-session-token"
	_ = db.CreateUserSession(context.Background(), user.ID, sessionKey)
	reqTok.AddCookie(&http.Cookie{Name: "session", Value: sessionKey})
	wTok := httptest.NewRecorder()
	h.GenerateKoboAuthToken(wTok, reqTok)
	if wTok.Code != http.StatusOK {
		t.Fatalf("Falha ao gerar Kobo token: %d %s", wTok.Code, wTok.Body.String())
	}
	var tokResp map[string]any
	_ = json.Unmarshal(wTok.Body.Bytes(), &tokResp)
	token, ok := tokResp["token"].(string)
	if !ok || token == "" {
		t.Fatalf("Token Kobo vazio retornado: %+v", tokResp)
	}

	// 6. Kobo Initialization
	reqInit := httptest.NewRequest(http.MethodGet, "/kobo/"+token+"/v1/initialization", nil)
	reqInit.SetPathValue("token", token)
	wInit := httptest.NewRecorder()
	h.KoboInitialization(wInit, reqInit)
	if wInit.Code != http.StatusOK {
		t.Fatalf("Falha na inicialização Kobo: %d", wInit.Code)
	}

	// 7. Kobo Profile
	reqProf := httptest.NewRequest(http.MethodGet, "/kobo/"+token+"/v1/user/profile", nil)
	reqProf.SetPathValue("token", token)
	wProf := httptest.NewRecorder()
	h.KoboProfile(wProf, reqProf)
	if wProf.Code != http.StatusOK {
		t.Fatalf("Falha no perfil Kobo: %d", wProf.Code)
	}

	// 8. Kobo Library Sync
	reqSync := httptest.NewRequest(http.MethodGet, "/kobo/"+token+"/v1/library/sync", nil)
	reqSync.SetPathValue("token", token)
	wSync := httptest.NewRecorder()
	h.KoboSync(wSync, reqSync)
	if wSync.Code != http.StatusOK {
		t.Fatalf("Falha no library sync Kobo: %d", wSync.Code)
	}
	var syncItems []map[string]any
	if err := json.Unmarshal(wSync.Body.Bytes(), &syncItems); err != nil {
		t.Fatalf("Resposta Kobo Sync inválida: %v", err)
	}
	if syncItems == nil {
		t.Fatal("Esperava array JSON de sincronização Kobo")
	}

	// 9. Kobo Reading State
	bodyKobo := `{"CurrentBookmark":{"ProgressPercent":0.85,"Location":"kobo-loc-42"},"Status":"Reading"}`
	reqState := httptest.NewRequest(http.MethodPut, "/kobo/"+token+"/v1/library/book-1/state", strings.NewReader(bodyKobo))
	reqState.SetPathValue("token", token)
	reqState.SetPathValue("book_uuid", "book-1")
	wState := httptest.NewRecorder()
	h.KoboReadingState(wState, reqState)
	if wState.Code != http.StatusOK {
		t.Fatalf("Falha ao salvar Kobo reading state: %d", wState.Code)
	}
}

func TestDuplicateResolutionAndDismissal(t *testing.T) {
	_, h, cleanup := setupTestDB(t)
	defer cleanup()

	// 1. Preview da resolução
	reqPrev := httptest.NewRequest(http.MethodPost, "/duplicates/preview-resolution", strings.NewReader(`{"strategy":"newest"}`))
	reqPrev.Header.Set("Content-Type", "application/json")
	wPrev := httptest.NewRecorder()
	h.PreviewResolution(wPrev, reqPrev)
	if wPrev.Code != http.StatusOK {
		t.Fatalf("PreviewResolution falhou: %d %s", wPrev.Code, wPrev.Body.String())
	}
	var prevResp map[string]any
	_ = json.Unmarshal(wPrev.Body.Bytes(), &prevResp)
	if prevResp["success"] != true {
		t.Fatalf("Esperado success=true no preview: %+v", prevResp)
	}

	// 2. Dismiss de duplicado
	reqDismiss := httptest.NewRequest(http.MethodPost, "/duplicates/dismiss/grp_test_hash", nil)
	reqDismiss.SetPathValue("group_hash", "grp_test_hash")
	wDismiss := httptest.NewRecorder()
	h.DismissDuplicate(wDismiss, reqDismiss)
	if wDismiss.Code != http.StatusOK {
		t.Fatalf("DismissDuplicate falhou: %d", wDismiss.Code)
	}

	// 3. Undismiss de duplicado
	reqUndismiss := httptest.NewRequest(http.MethodPost, "/duplicates/undismiss/grp_test_hash", nil)
	reqUndismiss.SetPathValue("group_hash", "grp_test_hash")
	wUndismiss := httptest.NewRecorder()
	h.DismissDuplicate(wUndismiss, reqUndismiss)
	if wUndismiss.Code != http.StatusOK {
		t.Fatalf("UndismissDuplicate falhou: %d", wUndismiss.Code)
	}

	// 4. Execução da resolução
	reqExec := httptest.NewRequest(http.MethodPost, "/duplicates/execute-resolution", strings.NewReader(`{"strategy":"newest"}`))
	reqExec.Header.Set("Content-Type", "application/json")
	wExec := httptest.NewRecorder()
	h.ExecuteResolution(wExec, reqExec)
	if wExec.Code != http.StatusOK {
		t.Fatalf("ExecuteResolution falhou: %d", wExec.Code)
	}
}

func TestCSVStatsExport(t *testing.T) {
	_, h, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/cwa-stats-export-csv/books", nil)
	req.SetPathValue("tab_name", "books")
	w := httptest.NewRecorder()
	h.CWAStatsExportCSV(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("CWAStatsExportCSV falhou: %d", w.Code)
	}
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/csv") {
		t.Fatalf("Esperado Content-Type text/csv, recebido: %s", contentType)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Total Books") || !strings.Contains(body, "Total Authors") {
		t.Fatalf("Conteúdo CSV inválido: %s", body)
	}
}

func TestMetadataProviderAndSearch(t *testing.T) {
	_, h, cleanup := setupTestDB(t)
	defer cleanup()

	// 1. Providers
	reqProv := httptest.NewRequest(http.MethodGet, "/metadata/provider", nil)
	wProv := httptest.NewRecorder()
	h.MetadataProvider(wProv, reqProv)
	if wProv.Code != http.StatusOK {
		t.Fatalf("MetadataProvider falhou: %d", wProv.Code)
	}
	var provs []map[string]any
	if err := json.Unmarshal(wProv.Body.Bytes(), &provs); err != nil || len(provs) < 2 {
		t.Fatalf("Esperado lista de provedores válida: %v", provs)
	}

	// 2. Search (sem query retorna array vazio)
	reqEmpty := httptest.NewRequest(http.MethodGet, "/metadata/search", nil)
	wEmpty := httptest.NewRecorder()
	h.MetadataSearch(wEmpty, reqEmpty)
	if wEmpty.Code != http.StatusOK {
		t.Fatalf("MetadataSearch vazio falhou: %d", wEmpty.Code)
	}
}

func TestBackupLifecycle(t *testing.T) {
	_, h, cleanup := setupTestDB(t)
	defer cleanup()

	// 1. Criar backup
	reqCreate := httptest.NewRequest(http.MethodPost, "/cwa-backup/create", nil)
	wCreate := httptest.NewRecorder()
	h.CWABackupCreate(wCreate, reqCreate)
	if wCreate.Code != http.StatusOK {
		t.Fatalf("CWABackupCreate falhou: %d - %s", wCreate.Code, wCreate.Body.String())
	}

	var resCreate map[string]any
	if err := json.Unmarshal(wCreate.Body.Bytes(), &resCreate); err != nil {
		t.Fatalf("JSON de backup inválido: %v", err)
	}
	filename, _ := resCreate["filename"].(string)
	if filename == "" {
		t.Fatal("Filename não retornado na criação do backup")
	}

	// 2. Listar backups
	reqList := httptest.NewRequest(http.MethodGet, "/cwa-backup/list", nil)
	wList := httptest.NewRecorder()
	h.CWABackupList(wList, reqList)
	if wList.Code != http.StatusOK {
		t.Fatalf("CWABackupList falhou: %d", wList.Code)
	}

	// 3. Download do backup
	reqDown := httptest.NewRequest(http.MethodGet, "/cwa-backup/download/"+filename, nil)
	reqDown.SetPathValue("filename", filename)
	wDown := httptest.NewRecorder()
	h.CWABackupDownload(wDown, reqDown)
	if wDown.Code != http.StatusOK {
		t.Fatalf("CWABackupDownload falhou: %d", wDown.Code)
	}
	if wDown.Header().Get("Content-Type") != "application/zip" {
		t.Fatalf("Esperado Content-Type application/zip, obtido %s", wDown.Header().Get("Content-Type"))
	}

	// 4. Apagar backup
	reqDel := httptest.NewRequest(http.MethodPost, "/cwa-backup/delete/"+filename, nil)
	reqDel.SetPathValue("filename", filename)
	wDel := httptest.NewRecorder()
	h.CWABackupDelete(wDel, reqDel)
	if wDel.Code != http.StatusOK {
		t.Fatalf("CWABackupDelete falhou: %d", wDel.Code)
	}
}

func TestAjaxCheckUpdate(t *testing.T) {
	_, h, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/ajax/check-update", nil)
	w := httptest.NewRecorder()
	h.AjaxCheckUpdate(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("AjaxCheckUpdate falhou: %d", w.Code)
	}
	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("Resposta JSON inválida do updater: %v", err)
	}
	if _, ok := res["current_version"]; !ok {
		t.Fatal("Esperado campo current_version na resposta")
	}
}

func TestHardcoverHandlers(t *testing.T) {
	_, h, cleanup := setupTestDB(t)
	defer cleanup()

	// 1. Iniciar processamento geral
	reqAll := httptest.NewRequest(http.MethodPost, "/cwa-hardcover-fetch/start", nil)
	wAll := httptest.NewRecorder()
	h.CWAHardcoverFetchAll(wAll, reqAll)
	if wAll.Code != http.StatusOK {
		t.Fatalf("CWAHardcoverFetchAll falhou: %d", wAll.Code)
	}

	// 2. Livro único
	reqSingle := httptest.NewRequest(http.MethodPost, "/metadata/hardcover/1", nil)
	reqSingle.SetPathValue("book_id", "1")
	wSingle := httptest.NewRecorder()
	h.MetadataHardcoverFetch(wSingle, reqSingle)
	// Como o livro 1 pode não existir na base limpa, espera 404 gracioso ou status not_found
	if wSingle.Code != http.StatusOK && wSingle.Code != http.StatusNotFound {
		t.Fatalf("MetadataHardcoverFetch código inesperado: %d", wSingle.Code)
	}
}

func TestOAuthLoginFlow(t *testing.T) {
	_, h, cleanup := setupTestDB(t)
	defer cleanup()

	// 1. Inicia login Google
	reqGoogle := httptest.NewRequest(http.MethodGet, "/login/google", nil)
	reqGoogle.SetPathValue("provider", "google")
	wGoogle := httptest.NewRecorder()
	h.OAuthLogin(wGoogle, reqGoogle)

	if wGoogle.Code != http.StatusFound {
		t.Fatalf("OAuthLogin esperava redirecionamento 302, obteve %d", wGoogle.Code)
	}
	loc := wGoogle.Header().Get("Location")
	if !strings.Contains(loc, "accounts.google.com") {
		t.Fatalf("Redirecionamento Google incorreto: %s", loc)
	}
	cookies := wGoogle.Result().Cookies()
	hasStateCookie := false
	for _, c := range cookies {
		if c.Name == "cwa_oauth_state" && c.Value != "" {
			hasStateCookie = true
			break
		}
	}
	if !hasStateCookie {
		t.Fatal("Cookie cwa_oauth_state não foi definido")
	}

	// 2. Callback com state inválido deve falhar com 400
	reqBadState := httptest.NewRequest(http.MethodGet, "/oauth/google/callback?state=invalid&code=abc", nil)
	reqBadState.SetPathValue("provider", "google")
	wBadState := httptest.NewRecorder()
	h.OAuthCallback(wBadState, reqBadState)
	if wBadState.Code != http.StatusBadRequest {
		t.Fatalf("OAuthCallback esperava erro 400 para CSRF inválido, obteve: %d", wBadState.Code)
	}
}


