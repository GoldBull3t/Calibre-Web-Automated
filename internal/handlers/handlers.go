package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/crocodilestick/calibre-web-go/internal/auth"
	"github.com/crocodilestick/calibre-web-go/internal/cache"
	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/crocodilestick/calibre-web-go/internal/mam"
	"github.com/crocodilestick/calibre-web-go/internal/plugins"
	"github.com/crocodilestick/calibre-web-go/internal/search"
	"github.com/flosch/pongo2/v6"
)

func init() {
	_ = pongo2.RegisterFilter("filesizeformat", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		size := in.Integer()
		if size < 1024 {
			return pongo2.AsValue(fmt.Sprintf("%d B", size)), nil
		} else if size < 1024*1024 {
			return pongo2.AsValue(fmt.Sprintf("%.1f KB", float64(size)/1024)), nil
		} else {
			return pongo2.AsValue(fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))), nil
		}
	})

	_ = pongo2.RegisterFilter("shortentitle", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		s := in.String()
		maxLen := 40
		if param != nil && param.Integer() > 0 {
			maxLen = param.Integer()
		}
		if len(s) > maxLen {
			return pongo2.AsValue(s[:maxLen] + "..."), nil
		}
		return pongo2.AsValue(s), nil
	})

	_ = pongo2.RegisterFilter("urlencode", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		return pongo2.AsValue(url.QueryEscape(in.String())), nil
	})
}

type Handler struct {
	db             *database.DB
	libraryDir     string // metadata.db, app.db, backups
	booksDir       string // books storage (Split Libraries)
	staticDir      string
	ingestDir      string
	pluginsDir     string
	searchIdx      *search.Indexer
	mamClient      *mam.Client
	pluginMgr      *plugins.Manager
	coverCache     *cache.ByteLRUCache
	gridTpl        *pongo2.Template
	detailTpl      *pongo2.Template
	modalTpl       *pongo2.Template
	readerTpl      *pongo2.Template
	entitiesTpl    *pongo2.Template
	searchFormTpl  *pongo2.Template
	adminTpl       *pongo2.Template
	tasksTpl       *pongo2.Template
	statsTpl       *pongo2.Template
	duplicatesTpl  *pongo2.Template
	loginTpl       *pongo2.Template
	bookEditTpl    *pongo2.Template
	shelfEditTpl   *pongo2.Template
	cwaSettingsTpl *pongo2.Template
	cwaConvertTpl  *pongo2.Template
	cwaFixerTpl    *pongo2.Template
	cwaBackupTpl   *pongo2.Template
}

func New(db *database.DB, libraryDir, booksDir, staticDir, ingestDir, pluginsDir string) (*Handler, error) {
	if booksDir == "" {
		booksDir = libraryDir
	}
	if pluginsDir == "" {
		pluginsDir = filepath.Join(libraryDir, "plugins")
	}

	h := &Handler{
		db:         db,
		libraryDir: libraryDir,
		booksDir:   booksDir,
		staticDir:  staticDir,
		ingestDir:  ingestDir,
		pluginsDir: pluginsDir,
		coverCache: cache.NewByteLRUCache(8 * 1024 * 1024), // 8 MB LRU em memória para capas
	}

	if db.AppDB() != nil {
		h.searchIdx, _ = search.NewIndexer(db.AppDB(), booksDir)
	}
	h.mamClient = mam.NewClient(os.Getenv("MAM_ID"))
	h.pluginMgr, _ = plugins.NewManager(pluginsDir)

	var err error
	h.gridTpl, err = pongo2.FromString(gridHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar grid template: %w", err)
	}

	h.detailTpl, err = pongo2.FromString(detailHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar detail template: %w", err)
	}

	h.modalTpl, err = pongo2.FromString(modalHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar modal template: %w", err)
	}

	h.readerTpl, err = pongo2.FromString(readerHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar reader template: %w", err)
	}

	h.entitiesTpl, err = pongo2.FromString(entitiesHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar entities template: %w", err)
	}

	h.searchFormTpl, err = pongo2.FromString(searchFormHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar searchForm template: %w", err)
	}

	h.adminTpl, err = pongo2.FromString(adminHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar admin template: %w", err)
	}

	h.tasksTpl, err = pongo2.FromString(tasksHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar tasks template: %w", err)
	}

	h.statsTpl, err = pongo2.FromString(statsHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar stats template: %w", err)
	}

	h.duplicatesTpl, err = pongo2.FromString(duplicatesHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar duplicates template: %w", err)
	}

	h.loginTpl, err = pongo2.FromString(loginHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar login template: %w", err)
	}

	h.bookEditTpl, err = pongo2.FromString(bookEditHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar bookEdit template: %w", err)
	}

	h.shelfEditTpl, err = pongo2.FromString(shelfEditHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar shelfEdit template: %w", err)
	}

	h.cwaSettingsTpl, err = pongo2.FromString(cwaSettingsHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar cwaSettings template: %w", err)
	}

	h.cwaConvertTpl, err = pongo2.FromString(cwaConvertHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar cwaConvert template: %w", err)
	}

	h.cwaFixerTpl, err = pongo2.FromString(cwaFixerHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar cwaFixer template: %w", err)
	}

	h.cwaBackupTpl, err = pongo2.FromString(cwaBackupHTML)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar cwaBackup template: %w", err)
	}

	return h, nil
}

// ServeStatic serve ficheiros estáticos diretamente do disco com cacheamento seguro.
func (h *Handler) ServeStatic(w http.ResponseWriter, r *http.Request) {
	relPath := strings.TrimPrefix(r.URL.Path, "/static/")
	fullPath := filepath.Join(h.staticDir, filepath.FromSlash(relPath))

	fi, err := os.Stat(fullPath)
	if err != nil || fi.IsDir() {
		// Aliases comuns de compatibilidade para leitores EPUB
		if relPath == "js/libs/jszip.min.js" || relPath == "js/libs/jszip.js" {
			fallbackZip := filepath.Join(h.staticDir, "js", "compress", "jszip_epub.min.js")
			if fiZ, errZ := os.Stat(fallbackZip); errZ == nil && !fiZ.IsDir() {
				w.Header().Set("Cache-Control", "public, max-age=86400")
				w.Header().Set("Content-Type", "application/javascript")
				http.ServeFile(w, r, fallbackZip)
				return
			}
		}
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	http.ServeFile(w, r, fullPath)
}

// RobotsTxt entrega robots.txt padrão.
func (h *Handler) RobotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("User-agent: *\nDisallow: /admin/\nDisallow: /download/\n"))
}

// CurrentUser recupera o utilizador autenticado a partir do contexto da requisição ou cookie de sessão.
func (h *Handler) CurrentUser(r *http.Request) *database.User {
	if u := auth.UserFromContext(r.Context()); u != nil {
		return u
	}
	if cookie, err := r.Cookie("session"); err == nil && cookie.Value != "" {
		if u, err := h.db.GetUserBySessionKey(r.Context(), cookie.Value); err == nil && u != nil {
			return u
		}
	}
	return nil
}

// AuthMiddleware injeta o utilizador na requisição através da validação da sessão no cookie.
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var currentUser *database.User
		if cookie, err := r.Cookie("session"); err == nil && cookie.Value != "" {
			if u, err := h.db.GetUserBySessionKey(r.Context(), cookie.Value); err == nil && u != nil {
				currentUser = u
			}
		}

		if currentUser == nil {
			if guest, err := h.db.GetUserByName(r.Context(), "Guest"); err == nil && guest != nil {
				currentUser = guest
			}
		}

		ctx := r.Context()
		if currentUser != nil {
			ctx = auth.WithUser(ctx, currentUser)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
