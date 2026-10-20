package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/auth"
	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/crocodilestick/calibre-web-go/internal/events"
	"github.com/crocodilestick/calibre-web-go/internal/handlers"
	"github.com/crocodilestick/calibre-web-go/internal/spa"
	"github.com/crocodilestick/calibre-web-go/internal/watcher"
)

func main() {
	// Restrição de Memória: teto suave de 35MB para a heap de Go (mantém RSS total < 40MB)
	debug.SetMemoryLimit(35 * 1024 * 1024)
	debug.SetGCPercent(60)

	port := flag.String("port", getEnv("PORT", "8083"), "Porta do servidor HTTP")
	libraryDir := flag.String("library-path", getEnv("LIBRARY_PATH", "./library"), "Caminho da biblioteca Calibre (base padrão)")
	metadataDir := flag.String("metadata-path", getEnv("METADATA_PATH", ""), "Caminho da pasta de metadados e base de dados (Split Libraries)")
	booksDir := flag.String("books-path", getEnv("BOOKS_PATH", ""), "Caminho da pasta onde os livros são armazenados (Split Libraries)")
	pluginsDir := flag.String("plugins-path", getEnv("PLUGINS_PATH", "./plugins"), "Caminho dos plugins do Calibre")
	ingestDir := flag.String("ingest-path", getEnv("INGEST_PATH", "./ingest"), "Caminho da pasta de ingestão contínua")
	staticDir := flag.String("static-path", getEnv("STATIC_PATH", "./cps/static"), "Caminho dos assets estáticos")
	flag.Parse()

	// Split Libraries fallback
	if *metadataDir == "" {
		*metadataDir = *libraryDir
	}
	if *booksDir == "" {
		*booksDir = *libraryDir
	}

	// Assegura existência das pastas necessárias
	for _, p := range []string{*metadataDir, *booksDir, *ingestDir, *pluginsDir} {
		if err := os.MkdirAll(p, 0755); err != nil {
			log.Fatalf("[FATAL] Erro ao criar diretório %s: %v", p, err)
		}
	}

	dbPath := filepath.Join(*metadataDir, "metadata.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		emptyDB := filepath.Join("empty_library", "metadata.db")
		if _, errEmpty := os.Stat(emptyDB); errEmpty == nil {
			log.Printf("[INIT] Inicializando metadata.db a partir do template limpo: %s", emptyDB)
			if errCopy := copyFile(emptyDB, dbPath); errCopy != nil {
				log.Fatalf("[FATAL] Falha ao copiar metadata.db template: %v", errCopy)
			}
		} else {
			log.Fatalf("[FATAL] metadata.db não encontrado em: %s", dbPath)
		}
	}

	appDbPath := filepath.Join(*metadataDir, "app.db")
	if _, err := os.Stat(appDbPath); os.IsNotExist(err) {
		emptyAppDB := filepath.Join("empty_library", "app.db")
		if _, errEmpty := os.Stat(emptyAppDB); errEmpty == nil {
			log.Printf("[INIT] Inicializando app.db a partir do template limpo: %s", emptyAppDB)
			_ = copyFile(emptyAppDB, appDbPath)
		}
	}

	log.Printf("[INIT] A ligar ao Calibre metadata.db em %s", dbPath)
	log.Printf("[INIT] Pasta de armazenamento de livros (Split Libraries): %s", *booksDir)
	db, err := database.Open(dbPath, appDbPath)
	if err != nil {
		log.Fatalf("[FATAL] Erro ao conectar ao SQLite: %v", err)
	}
	defer db.Close()

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Inicia motor de monitorização fsnotify na pasta de livros configurada
	w := watcher.New(*ingestDir, *booksDir, db)
	if err := w.Start(rootCtx); err != nil {
		log.Printf("[WARN] Não foi possível iniciar o watcher na pasta %s: %v", *ingestDir, err)
	} else {
		log.Printf("[WATCHER] Ingestão automática ativa na pasta: %s", *ingestDir)
	}

	// Inicializa handlers
	h, err := handlers.New(db, *metadataDir, *booksDir, *staticDir, *ingestDir, *pluginsDir)
	if err != nil {
		log.Fatalf("[FATAL] Falha ao inicializar handlers web: %v", err)
	}

	spaHandler, _ := spa.New()

	mux := http.NewServeMux()

	// 1. Navegação e listagens principais
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("classic") == "1" || !spaHandler.HasAssets() {
			h.ListBooks(w, r)
			return
		}
		spaHandler.ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /classic", h.ListBooks)
	if spaHandler.HasAssets() {
		mux.Handle("GET /assets/", spaHandler)
		mux.Handle("GET /manifest.webmanifest", spaHandler)
		mux.Handle("GET /manifest.json", spaHandler)
		mux.Handle("GET /sw.js", spaHandler)
		mux.Handle("GET /favicon.svg", spaHandler)
		mux.Handle("GET /app", spaHandler)
		mux.Handle("GET /app/", spaHandler)
	}
	mux.HandleFunc("GET /page/{page}", h.ListBooks)

	dataTypes := []string{
		"author", "series", "category", "tags", "publisher",
		"ratings", "formats", "language", "hot", "discover",
		"newest", "readbooks", "unreadbooks",
	}
	for _, d := range dataTypes {
		dataType := d
		mux.HandleFunc("GET /"+dataType+"/{sort_param}", func(w http.ResponseWriter, r *http.Request) {
			r.SetPathValue("data", dataType)
			h.BrowseUniversal(w, r)
		})
		mux.HandleFunc("GET /"+dataType+"/{sort_param}/{book_id}", func(w http.ResponseWriter, r *http.Request) {
			r.SetPathValue("data", dataType)
			h.BrowseUniversal(w, r)
		})
		mux.HandleFunc("GET /"+dataType+"/{sort_param}/{book_id}/{page}", func(w http.ResponseWriter, r *http.Request) {
			r.SetPathValue("data", dataType)
			h.BrowseUniversal(w, r)
		})
	}

	// Aliases de navegação direta
	mux.HandleFunc("GET /authors", h.ListAuthors)
	mux.HandleFunc("GET /author", h.ListAuthors)
	mux.HandleFunc("GET /series", h.ListSeries)
	mux.HandleFunc("GET /tags", h.ListTags)
	mux.HandleFunc("GET /category", h.ListTags)
	mux.HandleFunc("GET /formats", h.ListFormats)
	mux.HandleFunc("GET /publishers", h.ListPublishers)
	mux.HandleFunc("GET /publisher", h.ListPublishers)
	mux.HandleFunc("GET /ratings", h.ListRatings)
	mux.HandleFunc("GET /languages", h.ListLanguages)
	mux.HandleFunc("GET /language", h.ListLanguages)
	mux.HandleFunc("GET /hot", h.ListBooks)
	mux.HandleFunc("GET /discover", h.ListBooks)
	mux.HandleFunc("GET /newest", h.ListBooks)
	mux.HandleFunc("GET /readbooks", h.ListBooks)
	mux.HandleFunc("GET /unreadbooks", h.ListBooks)
	mux.HandleFunc("GET /shelf/create", h.CreateShelf)
	mux.HandleFunc("POST /shelf/create", h.CreateShelf)
	mux.HandleFunc("GET /shelf/edit/{id}", h.EditShelf)
	mux.HandleFunc("POST /shelf/edit/{id}", h.EditShelf)
	mux.HandleFunc("POST /shelf/delete/{id}", h.DeleteShelf)
	mux.HandleFunc("POST /shelf/add/{shelf_id}/{book_id}", h.AddToShelf)
	mux.HandleFunc("POST /shelf/remove/{shelf_id}/{book_id}", h.RemoveFromShelf)
	mux.HandleFunc("POST /shelf/add_selected_to_shelf", h.AddSelectedToShelf)
	mux.HandleFunc("POST /shelf/massadd/{shelf_id}", h.MassAddToShelf)
	mux.HandleFunc("GET /shelf/order/{shelf_id}", h.OrderShelf)
	mux.HandleFunc("POST /shelf/order/{shelf_id}", h.OrderShelf)
	mux.HandleFunc("GET /shelf/{id}", func(w http.ResponseWriter, r *http.Request) {
		r.SetPathValue("data", "shelf")
		r.SetPathValue("book_id", r.PathValue("id"))
		h.BrowseUniversal(w, r)
	})
	mux.HandleFunc("GET /simpleshelf/{id}", func(w http.ResponseWriter, r *http.Request) {
		r.SetPathValue("data", "shelf")
		r.SetPathValue("book_id", r.PathValue("id"))
		h.BrowseUniversal(w, r)
	})
	mux.HandleFunc("GET /magicshelf/{id}", func(w http.ResponseWriter, r *http.Request) {
		r.SetPathValue("data", "shelf")
		r.SetPathValue("book_id", r.PathValue("id"))
		h.BrowseUniversal(w, r)
	})

	// 2. Livro: detalhes, edição, capas, download e streaming
	mux.HandleFunc("GET /book/{id}", h.ShowBook)
	mux.HandleFunc("GET /admin/book/{id}", h.ShowEditBook)
	mux.HandleFunc("POST /admin/book/{id}", h.EditBook)
	mux.HandleFunc("POST /admin/book/convert/{id}", h.ConvertBook)
	mux.HandleFunc("POST /send/{id}/{format}/{convert}", h.SendBook)
	mux.HandleFunc("POST /send_selected/{id}", h.SendBook)
	mux.HandleFunc("GET /cover/{id}", h.GetCover)
	mux.HandleFunc("GET /cover/{id}/{resolution}", h.GetCover)
	mux.HandleFunc("GET /series_cover/{id}", h.GetSeriesCover)
	mux.HandleFunc("GET /series_cover/{id}/{resolution}", h.GetSeriesCover)
	mux.HandleFunc("GET /download/{id}", h.DownloadBook)
	mux.HandleFunc("GET /download/{id}/{format}", h.DownloadBook)
	mux.HandleFunc("GET /download/{id}/{format}/{name}", h.DownloadBook)
	mux.HandleFunc("GET /show/{id}/{format}", h.ShowBookStream)
	mux.HandleFunc("GET /show/{id}/{format}/{name}", h.ShowBookStream)
	mux.HandleFunc("GET /read/{id}", h.ReaderWeb)
	mux.HandleFunc("GET /read/{id}/{format}", h.ReaderWeb)
	mux.HandleFunc("GET /api/comic/{id}/pages", h.APIComicPages)
	mux.HandleFunc("GET /api/comic/{id}/page/{page}", h.APIComicPageImage)

	// 3. Operações e AJAX
	mux.HandleFunc("POST /upload", h.UploadBook)
	mux.HandleFunc("POST /edit-book/upload", h.UploadBook)
	mux.HandleFunc("POST /delete/{id}", h.DeleteBook)
	mux.HandleFunc("POST /delete/{id}/{format}", h.DeleteBookFormat)
	mux.HandleFunc("POST /ajax/delete/{id}", h.DeleteBook)
	mux.HandleFunc("POST /edit-book/delete/{id}", h.DeleteBook)
	mux.HandleFunc("POST /ajax/toggleread/{id}", h.ToggleRead)
	mux.HandleFunc("POST /ajax/togglearchived/{id}", h.ToggleArchived)
	mux.HandleFunc("POST /ajax/bookmark/{id}/{format}", h.Bookmark)
	mux.HandleFunc("GET /ajax/emailstat", h.EmailStat)
	mux.HandleFunc("GET /ajax/listbooks", h.ListBooksJSON)
	mux.HandleFunc("GET /table", h.ListBooksJSON)
	mux.HandleFunc("POST /ajax/displayselectedbooks", h.DisplaySelectedBooks)
	mux.HandleFunc("POST /ajax/archiveselectedbooks", h.ArchiveSelectedBooks)
	mux.HandleFunc("POST /ajax/deleteselectedbooks", h.DeleteSelectedBooks)
	mux.HandleFunc("POST /ajax/readselectedbooks", h.ReadSelectedBooks)
	mux.HandleFunc("POST /ajax/editselectedbooks", h.EditSelectedBooks)
	mux.HandleFunc("POST /ajax/mergebooks", h.MergeBooks)

	// 4. Pesquisas e Autocompletes
	mux.HandleFunc("GET /search", h.Search)
	mux.HandleFunc("GET /advsearch", h.AdvSearch)
	mux.HandleFunc("POST /advsearch", h.AdvSearch)
	mux.HandleFunc("GET /get_authors_json", h.GetAuthorsJSON)
	mux.HandleFunc("GET /get_publishers_json", h.GetPublishersJSON)
	mux.HandleFunc("GET /get_tags_json", h.GetTagsJSON)
	mux.HandleFunc("GET /get_series_json", h.GetSeriesJSON)
	mux.HandleFunc("GET /get_languages_json", h.GetLanguagesJSON)
	mux.HandleFunc("GET /get_matching_tags", h.GetMatchingTags)
	mux.HandleFunc("POST /metadata/search", h.MetadataSearch)
	mux.HandleFunc("GET /metadata/search", h.MetadataSearch)
	mux.HandleFunc("GET /metadata/provider", h.MetadataProvider)
	mux.HandleFunc("POST /metadata/provider", h.MetadataProviderChange)
	mux.HandleFunc("POST /metadata/provider/{prov_name}", h.MetadataProviderChange)

	// 5. Administração, Tarefas, Stats e CWA
	mux.HandleFunc("GET /admin", h.Admin)
	mux.HandleFunc("GET /admin/view", h.Admin)
	mux.HandleFunc("GET /admin/config", h.Admin)
	mux.HandleFunc("GET /admin/viewconfig", h.Admin)
	mux.HandleFunc("GET /admin/usertable", h.Admin)
	mux.HandleFunc("GET /tasks", h.Tasks)
	mux.HandleFunc("GET /stats", h.Stats)
	mux.HandleFunc("GET /cwa-stats", h.Stats)
	mux.HandleFunc("GET /cwa-stats-show", h.Stats)
	mux.HandleFunc("GET /cwa-stats-export-csv/{tab_name}", h.CWAStatsExportCSV)
	mux.HandleFunc("GET /cwa-stats-debug", h.CWAStatsDebug)
	mux.HandleFunc("GET /cwa-logs/read/{log_filename}", h.CWALogRead)
	mux.HandleFunc("GET /cwa-logs/download/{log_filename}", h.CWALogDownload)
	mux.HandleFunc("GET /duplicates", h.Duplicates)
	mux.HandleFunc("POST /duplicates/trigger-scan", h.TriggerDuplicateScan)
	mux.HandleFunc("POST /duplicates/dismiss-setup-notice", h.DismissDuplicateSetupNotice)
	mux.HandleFunc("POST /duplicates/dismiss/{group_hash}", h.DismissDuplicate)
	mux.HandleFunc("POST /duplicates/undismiss/{group_hash}", h.DismissDuplicate)
	mux.HandleFunc("POST /duplicates/preview-resolution", h.PreviewResolution)
	mux.HandleFunc("POST /duplicates/execute-resolution", h.ExecuteResolution)
	mux.HandleFunc("GET /duplicates/status", h.DuplicateStatus)
	mux.HandleFunc("GET /cwa-settings", h.CWASettings)
	mux.HandleFunc("GET /cwa-switch-theme", h.CWASwitchTheme)
	mux.HandleFunc("POST /cwa-switch-theme", h.CWASwitchTheme)
	mux.HandleFunc("POST /cwa-library-refresh", h.LibraryRefresh)
	mux.HandleFunc("GET /cwa-library-refresh/messages", h.LibraryRefreshMessages)
	mux.HandleFunc("GET /cwa-check-monitoring", h.CWACheckMonitoring)
	mux.HandleFunc("GET /cwa-scheduled/upcoming", h.CWAScheduledUpcoming)
	mux.HandleFunc("GET /cwa-scheduled/upcoming-ops", h.CWAScheduledUpcomingOps)
	mux.HandleFunc("POST /cwa-scheduled/cancel", h.CWAScheduledCancel)
	mux.HandleFunc("GET /cwa-convert-library-overview", h.CWAConvertLibraryOverview)
	mux.HandleFunc("GET /cwa-convert-library-start", h.CWAConvertLibraryStart)
	mux.HandleFunc("GET /convert-library-cancel", h.ConvertLibraryCancel)
	mux.HandleFunc("GET /convert-library-status", h.ConvertLibraryStatus)
	mux.HandleFunc("GET /cwa-convert-library/schedule/{delay}", h.CWAConvertLibraryStart)
	mux.HandleFunc("GET /cwa-convert-library/log-archive", h.CWAConvertLibraryOverview)
	mux.HandleFunc("GET /cwa-epub-fixer-overview", h.CWAEpubFixerOverview)
	mux.HandleFunc("GET /cwa-epub-fixer-start", h.CWAEpubFixerStart)
	mux.HandleFunc("GET /epub-fixer-cancel", h.EpubFixerCancel)
	mux.HandleFunc("GET /epub-fixer-status", h.EpubFixerStatus)
	mux.HandleFunc("GET /cwa-epub-fixer/schedule/{delay}", h.CWAEpubFixerStart)
	mux.HandleFunc("GET /cwa-epub-fixer/log-archive", h.CWAEpubFixerOverview)
	mux.HandleFunc("POST /cwa-epub-fixer/run-book", h.CWAEpubFixerStart)
	mux.HandleFunc("GET /cwa-backup/list", h.CWABackupList)
	mux.HandleFunc("POST /cwa-backup/create", h.CWABackupCreate)
	mux.HandleFunc("GET /cwa-backup/download/{filename}", h.CWABackupDownload)
	mux.HandleFunc("POST /cwa-backup/delete/{filename}", h.CWABackupDelete)
	mux.HandleFunc("GET /ajax/check-update", h.AjaxCheckUpdate)
	mux.HandleFunc("GET /cwa-check-update", h.AjaxCheckUpdate)
	mux.HandleFunc("POST /metadata/hardcover/{book_id}", h.MetadataHardcoverFetch)
	mux.HandleFunc("POST /cwa-hardcover-fetch/start", h.CWAHardcoverFetchAll)

	// 6. Autenticação e Perfis (com Rate Limiter Token Bucket para prevenção de força bruta)
	authLimiter := auth.NewRateLimiter(5, 10) // 5 req/s, capacidade máxima de burst 10
	mux.HandleFunc("GET /login", h.Login)
	mux.Handle("POST /login", authLimiter.Middleware(http.HandlerFunc(h.Login)))
	mux.HandleFunc("GET /login/{provider}", h.OAuthLogin)
	mux.HandleFunc("GET /oauth/{provider}/callback", h.OAuthCallback)
	mux.HandleFunc("GET /oauth/callback", h.OAuthCallback)
	mux.HandleFunc("GET /login/{provider}/authorized", h.OAuthCallback)
	mux.HandleFunc("GET /logout", h.Logout)
	mux.HandleFunc("POST /logout", h.Logout)
	mux.HandleFunc("GET /register", h.Register)
	mux.Handle("POST /register", authLimiter.Middleware(http.HandlerFunc(h.Register)))
	mux.HandleFunc("GET /me", h.Profile)
	mux.HandleFunc("POST /me", h.Profile)
	mux.HandleFunc("GET /me/profile-picture", h.ProfilePicture)
	mux.HandleFunc("GET /user_profiles.json", h.UserProfilesJSON)

	// 7. KOReader Sync & Kobo Sync
	mux.HandleFunc("/kosync", h.KOSync)
	mux.HandleFunc("/kosync/{path...}", h.KOSync)
	mux.HandleFunc("GET /users/auth", h.KOSyncAuth)
	mux.HandleFunc("POST /users/create", h.KOSyncCreateUser)
	mux.HandleFunc("GET /syncs/progress/{document...}", h.KOSyncGetProgress)
	mux.HandleFunc("PUT /syncs/progress", h.KOSyncUpdateProgress)

	// Kobo Sync e Kobo Auth
	mux.HandleFunc("GET /kobo/{token}/v1/initialization", h.KoboInitialization)
	mux.HandleFunc("GET /v1/initialization", h.KoboInitialization)
	mux.HandleFunc("GET /kobo/{token}/v1/user/profile", h.KoboProfile)
	mux.HandleFunc("GET /v1/user/profile", h.KoboProfile)
	mux.HandleFunc("GET /kobo/{token}/v1/analytics/gettests", h.KoboAnalytics)
	mux.HandleFunc("GET /v1/analytics/gettests", h.KoboAnalytics)
	mux.HandleFunc("GET /kobo/{token}/v1/library/sync", h.KoboSync)
	mux.HandleFunc("GET /v1/library/sync", h.KoboSync)
	mux.HandleFunc("GET /kobo/{token}/download/{id}/{format}", h.KoboDownload)
	mux.HandleFunc("GET /kobo/{token}/{book_uuid}/{width}/{height}/{is_greyscale}/image.jpg", h.KoboCover)
	mux.HandleFunc("GET /kobo/{token}/{book_uuid}/{width}/{height}/{quality}/{is_greyscale}/image.jpg", h.KoboCover)
	mux.HandleFunc("PUT /kobo/{token}/v1/library/{book_uuid}/state", h.KoboReadingState)
	mux.HandleFunc("GET /kobo_auth/generate_auth_token/{user_id}", h.GenerateKoboAuthToken)
	mux.HandleFunc("POST /kobo_auth/generate_auth_token/{user_id}", h.GenerateKoboAuthToken)

	// 8. Catálogo OPDS 1.2
	mux.HandleFunc("GET /opds", h.OPDSFeed)
	mux.HandleFunc("GET /opds/", h.OPDSFeed)
	mux.HandleFunc("GET /opds/new", h.OPDSFeed)
	mux.HandleFunc("GET /opds/hot", h.OPDSFeed)
	mux.HandleFunc("GET /opds/discover", h.OPDSFeed)
	mux.HandleFunc("GET /opds/author", h.OPDSFeed)
	mux.HandleFunc("GET /opds/author/{id}", h.OPDSFeed)
	mux.HandleFunc("GET /opds/publisher", h.OPDSFeed)
	mux.HandleFunc("GET /opds/publisher/{id}", h.OPDSFeed)
	mux.HandleFunc("GET /opds/series", h.OPDSFeed)
	mux.HandleFunc("GET /opds/series/{id}", h.OPDSFeed)
	mux.HandleFunc("GET /opds/category", h.OPDSFeed)
	mux.HandleFunc("GET /opds/category/{id}", h.OPDSFeed)
	mux.HandleFunc("GET /opds/ratings", h.OPDSFeed)
	mux.HandleFunc("GET /opds/ratings/{id}", h.OPDSFeed)
	mux.HandleFunc("GET /opds/formats", h.OPDSFeed)
	mux.HandleFunc("GET /opds/formats/{id}", h.OPDSFeed)
	mux.HandleFunc("GET /opds/language", h.OPDSFeed)
	mux.HandleFunc("GET /opds/language/{id}", h.OPDSFeed)
	mux.HandleFunc("GET /opds/readbooks", h.OPDSFeed)
	mux.HandleFunc("GET /opds/unreadbooks", h.OPDSFeed)
	mux.HandleFunc("GET /opds/search", h.OPDSFeed)
	mux.HandleFunc("GET /opds/search/{query}", h.OPDSFeed)
	mux.HandleFunc("GET /opds/osd", h.OPDSFeed)
	mux.HandleFunc("GET /opds/osd.xml", h.OPDSFeed)
	mux.HandleFunc("GET /opds/cover/{id}", h.GetCover)
	mux.HandleFunc("GET /opds/cover_240_240/{id}", h.GetCover)
	mux.HandleFunc("GET /opds/cover_90_90/{id}", h.GetCover)
	mux.HandleFunc("GET /opds/thumb_240_240/{id}", h.GetCover)
	mux.HandleFunc("GET /opds/download/{id}/{format}/", h.DownloadBook)

	// 9. NextGen APIs (Real-time SSE, Cover Upload, Progress Sync, FTS5 Search, MAM 🐭🧀, Plugins & DeDRM)
	mux.Handle("GET /api/events", events.GetBroker())
	mux.HandleFunc("POST /api/books/{id}/cover", h.UploadCover)
	mux.HandleFunc("GET /api/reader/progress/{id}", h.ReaderGetProgress)
	mux.HandleFunc("POST /api/reader/progress/{id}", h.ReaderSaveProgress)
	mux.HandleFunc("GET /api/books", h.APIGetBooks)
	mux.HandleFunc("GET /api/books/{id}", h.APIGetBook)
	mux.HandleFunc("GET /api/stats/overview", h.APIStatsOverview)
	mux.HandleFunc("GET /api/search/fts", h.SearchFTS)
	mux.HandleFunc("POST /api/search/fts/reindex", h.ReindexFTS)
	mux.HandleFunc("GET /api/mam/status", h.MAMStatus)
	mux.HandleFunc("POST /api/mam/config", h.MAMConfig)
	mux.HandleFunc("GET /api/mam/search", h.MAMSearch)
	mux.HandleFunc("POST /api/mam/download/{id}", h.MAMDownload)
	mux.HandleFunc("GET /api/admin/plugins", h.PluginsList)
	mux.HandleFunc("POST /api/admin/plugins/upload", h.PluginsUpload)
	mux.HandleFunc("POST /api/admin/plugins/toggle/{filename}", h.PluginsToggle)
	mux.HandleFunc("GET /api/authors", h.APIGetAuthors)
	mux.HandleFunc("GET /api/series", h.APIGetSeries)
	mux.HandleFunc("GET /api/tags", h.APIGetTags)
	mux.HandleFunc("GET /api/formats", h.APIGetFormats)
	mux.HandleFunc("GET /api/shelves", h.APIGetShelves)
	mux.HandleFunc("POST /api/shelves", h.APICreateShelf)
	mux.HandleFunc("DELETE /api/shelves/{id}", h.APIDeleteShelf)
	mux.HandleFunc("POST /api/books/batch", h.APIBatchBooks)
	mux.HandleFunc("POST /api/books/{id}/edit", h.APIEditBook)
	mux.HandleFunc("GET /api/duplicates", h.APIGetDuplicates)
	mux.HandleFunc("POST /api/duplicates/scan", h.TriggerDuplicateScan)
	mux.HandleFunc("POST /api/duplicates/resolve", h.ExecuteResolution)
	mux.HandleFunc("POST /api/duplicates/dismiss/{group_hash}", h.DismissDuplicate)
	mux.HandleFunc("GET /api/cwa/settings", h.APIGetCWASettings)
	mux.HandleFunc("POST /api/cwa/settings", h.APISaveCWASettings)
	mux.HandleFunc("GET /api/backups", h.APIGetBackups)
	mux.HandleFunc("POST /api/backups/create", h.APICreateBackup)
	mux.HandleFunc("DELETE /api/backups/{filename}", h.APIDeleteBackup)
	mux.HandleFunc("POST /api/fixer/start", h.CWAEpubFixerStart)
	mux.HandleFunc("GET /api/fixer/status", h.EpubFixerStatus)
	mux.HandleFunc("POST /api/convert/start", h.CWAConvertLibraryStart)
	mux.HandleFunc("GET /api/convert/status", h.ConvertLibraryStatus)

	// 9.1 Autenticação Moderna: Passkeys (WebAuthn) e OIDC / SSO
	mux.HandleFunc("GET /api/auth/passkey/register/options", h.APIPasskeyRegisterOptions)
	mux.HandleFunc("POST /api/auth/passkey/register/verify", h.APIPasskeyRegisterVerify)
	mux.HandleFunc("GET /api/auth/passkey/login/options", h.APIPasskeyLoginOptions)
	mux.Handle("POST /api/auth/passkey/login/verify", authLimiter.Middleware(http.HandlerFunc(h.APIPasskeyLoginVerify)))
	mux.HandleFunc("GET /api/auth/passkeys", h.APIListPasskeys)
	mux.HandleFunc("DELETE /api/auth/passkeys/{id}", h.APIDeletePasskey)
	mux.HandleFunc("GET /api/auth/oidc/config", h.APIGetOIDCConfig)
	mux.HandleFunc("POST /api/auth/oidc/config", h.APISaveOIDCConfig)
	mux.HandleFunc("GET /auth/oidc/login", h.OIDCLogin)
	mux.HandleFunc("GET /auth/oidc/callback", h.OIDCCallback)

	// 9.2 Auto-Enriquecimento Inteligente (Online Metadata, HD Covers & Gutenberg Cleaner)
	mux.HandleFunc("GET /api/books/{id}/enrich/search", h.APIEnrichSearch)
	mux.HandleFunc("POST /api/books/{id}/enrich/apply", h.APIEnrichApply)
	mux.HandleFunc("POST /api/books/{id}/enrich/auto", h.APIEnrichAuto)
	mux.HandleFunc("POST /api/enrich/clean-library", h.APICleanLibraryMetadata)

	// 10. Assets estáticos e utilitários
	mux.HandleFunc("GET /static/", h.ServeStatic)
	mux.HandleFunc("GET /robots.txt", h.RobotsTxt)

	// Endpoint de métricas de saúde
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":       "ok",
			"heap_alloc":   fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024),
			"heap_sys":     fmt.Sprintf("%.2f MB", float64(m.Sys)/1024/1024),
			"num_gc":       m.NumGC,
			"goroutines":   runtime.NumGoroutine(),
			"sqlite_cache": "2 MB (WAL mode)",
		})
	})

	srv := &http.Server{
		Addr:         ":" + *port,
		Handler:      h.AuthMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("[SERVER] Calibre-Web operacional em http://localhost:%s", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Erro no servidor HTTP: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("[SHUTDOWN] Interrupção detectada. A terminar processos graciosamente...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERROR] Erro ao encerrar servidor HTTP: %v", err)
	}

	log.Println("[SHUTDOWN] Servidor desligado com segurança.")
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
