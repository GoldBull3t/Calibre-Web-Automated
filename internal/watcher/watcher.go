package watcher

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/crocodilestick/calibre-web-go/internal/formats"
	"github.com/crocodilestick/calibre-web-go/internal/metadata"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

var supportedExtensions = map[string]bool{
	".epub": true,
	".pdf":  true,
	".mobi": true,
	".azw":  true,
	".azw3": true,
	".cbz":  true,
	".cbr":  true,
	".txt":  true,
	".fb2":  true,
	".zip":  true,
}

type Watcher struct {
	ingestDir  string
	libraryDir string
	db         *database.DB
	timers     map[string]*time.Timer
	mu         sync.Mutex
	jobs       chan string
}

func New(ingestDir, libraryDir string, db *database.DB) *Watcher {
	return &Watcher{
		ingestDir:  ingestDir,
		libraryDir: libraryDir,
		db:         db,
		timers:     make(map[string]*time.Timer),
		jobs:       make(chan string, 100),
	}
}

func (w *Watcher) Start(ctx context.Context) error {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("falha ao inicializar fsnotify: %w", err)
	}

	if err := os.MkdirAll(w.ingestDir, 0755); err != nil {
		return err
	}

	go w.worker(ctx)

	go func() {
		defer fsw.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-fsw.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					ext := strings.ToLower(filepath.Ext(event.Name))
					if supportedExtensions[ext] {
						w.debounce(event.Name)
					}
				}
			case err, ok := <-fsw.Errors:
				if !ok {
					return
				}
				log.Printf("[WATCHER] Erro no fsnotify: %v", err)
			}
		}
	}()

	return fsw.Add(w.ingestDir)
}

func (w *Watcher) debounce(filePath string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if t, exists := w.timers[filePath]; exists {
		t.Stop()
	}

	w.timers[filePath] = time.AfterFunc(2*time.Second, func() {
		w.mu.Lock()
		delete(w.timers, filePath)
		w.mu.Unlock()
		w.jobs <- filePath
	})
}

func (w *Watcher) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case filePath := <-w.jobs:
			ext := strings.ToLower(filepath.Ext(filePath))
			if ext == ".zip" {
				// Auto-Zipper: descompacta livros do ZIP e descarta o arquivo compactado
				extracted, err := formats.UnpackZip(filePath, w.ingestDir)
				_ = os.Remove(filePath)
				if err != nil {
					log.Printf("[AUTO-ZIPPER ERROR] Falha ao descompactar %s: %v", filePath, err)
				} else {
					log.Printf("[AUTO-ZIPPER] %d livros extraídos de %s", len(extracted), filePath)
					for _, ef := range extracted {
						w.debounce(ef)
					}
				}
				continue
			}

			if err := w.processFile(filePath); err != nil {
				log.Printf("[INGEST ERROR] Falha ao processar %s: %v", filePath, err)
			} else {
				log.Printf("[INGEST SUCCESS] Livro incorporado com sucesso: %s", filePath)
			}
		}
	}
}

// ProcessFile importa e normaliza um livro diretamente na biblioteca Calibre.
func (w *Watcher) ProcessFile(srcPath string) error {
	return w.processFile(srcPath)
}

func (w *Watcher) processFile(srcPath string) error {
	fi, err := os.Stat(srcPath)
	if err != nil || fi.IsDir() {
		return err
	}

	meta, err := formats.Extract(srcPath)
	if err != nil {
		return fmt.Errorf("metadados corrompidos ou ilegíveis: %w", err)
	}

	title := meta.Title
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	}
	author := "Unknown"
	if len(meta.Authors) > 0 && strings.TrimSpace(meta.Authors[0]) != "" {
		author = meta.Authors[0]
	}

	cleanAuthor := sanitizeName(author)
	cleanTitle := sanitizeName(title)

	return w.db.ExecuteWrite(func(tx *sql.Tx) error {
		// 1. Obter ou Criar Autor
		var authorID int64
		err := tx.QueryRow("SELECT id FROM authors WHERE name = ?", author).Scan(&authorID)
		if err == sql.ErrNoRows {
			res, err := tx.Exec("INSERT INTO authors (name, sort, link) VALUES (?, ?, '')", author, author)
			if err != nil {
				return fmt.Errorf("erro ao inserir autor: %w", err)
			}
			authorID, _ = res.LastInsertId()
		}

		// 2. Prevenção de duplicados na ingestão: verifica se livro já existe pelo título e autor
		var existingBookID int64
		var existingPath string
		err = tx.QueryRow(`
			SELECT b.id, b.path
			FROM books b
			JOIN books_authors_link bal ON b.id = bal.book
			WHERE b.title = ? COLLATE NOCASE AND bal.author = ?
		`, title, authorID).Scan(&existingBookID, &existingPath)

		if err == nil && existingBookID > 0 {
			// Livro já existe: anexa como novo formato em vez de duplicar
			var fmtCount int
			_ = tx.QueryRow("SELECT COUNT(*) FROM data WHERE book = ? AND format = ?", existingBookID, meta.Format).Scan(&fmtCount)
			if fmtCount == 0 {
				bookDir := filepath.Join(w.libraryDir, filepath.FromSlash(existingPath))
				finalFileName := fmt.Sprintf("%s - %s.%s", cleanTitle, cleanAuthor, strings.ToLower(meta.Format))
				finalPath := filepath.Join(bookDir, finalFileName)

				if err := moveOrCopy(srcPath, finalPath); err != nil {
					return fmt.Errorf("erro ao anexar formato à biblioteca: %w", err)
				}

				if len(meta.CoverBytes) > 0 {
					coverDest := filepath.Join(bookDir, "cover.jpg")
					if _, errC := os.Stat(coverDest); os.IsNotExist(errC) {
						if err := os.WriteFile(coverDest, meta.CoverBytes, 0644); err == nil {
							_, _ = tx.Exec("UPDATE books SET has_cover = 1 WHERE id = ?", existingBookID)
						}
					}
				}

				_, err = tx.Exec(`
					INSERT INTO data (book, format, uncompressed_size, name)
					VALUES (?, ?, ?, ?)
				`, existingBookID, meta.Format, fi.Size(), strings.TrimSuffix(finalFileName, "."+strings.ToLower(meta.Format)))
				return err
			}
			// Se o formato já existia, descarta o arquivo de ingestão duplicado
			_ = os.Remove(srcPath)
			return nil
		}

		// 3. Novo Livro
		bookUUID := uuid.New().String()
		now := time.Now().UTC()

		res, err := tx.Exec(`
			INSERT INTO books (title, sort, timestamp, pubdate, series_index, last_modified, path, has_cover, uuid)
			VALUES (?, ?, ?, ?, 1.0, ?, '', 0, ?)
		`, title, title, now, now, now, bookUUID)
		if err != nil {
			return fmt.Errorf("erro ao inserir livro: %w", err)
		}
		bookID, _ := res.LastInsertId()

		// 4. Vincular Livro ao Autor
		if _, err := tx.Exec("INSERT INTO books_authors_link (book, author) VALUES (?, ?)", bookID, authorID); err != nil {
			return fmt.Errorf("erro ao vincular autor: %w", err)
		}

		// 5. Estrutura de Diretórios Calibre: Autor/Título (ID)
		relPath := filepath.Join(cleanAuthor, fmt.Sprintf("%s (%d)", cleanTitle, bookID))
		absBookDir := filepath.Join(w.libraryDir, relPath)
		if err := os.MkdirAll(absBookDir, 0755); err != nil {
			return fmt.Errorf("erro ao criar pasta do livro: %w", err)
		}

		finalFileName := fmt.Sprintf("%s - %s.%s", cleanTitle, cleanAuthor, strings.ToLower(meta.Format))
		finalFilePath := filepath.Join(absBookDir, finalFileName)

		if err := moveOrCopy(srcPath, finalFilePath); err != nil {
			return fmt.Errorf("erro ao mover ficheiro do livro: %w", err)
		}

		// 6. Extração e Gravação de Capa
		hasCover := 0
		if len(meta.CoverBytes) > 0 {
			coverDest := filepath.Join(absBookDir, "cover.jpg")
			if err := os.WriteFile(coverDest, meta.CoverBytes, 0644); err == nil {
				hasCover = 1
			}
		}

		// 7. Inserir Formato em 'data'
		uncompressedSize := fi.Size()
		if _, err := tx.Exec(`
			INSERT INTO data (book, format, uncompressed_size, name)
			VALUES (?, ?, ?, ?)
		`, bookID, meta.Format, uncompressedSize, strings.TrimSuffix(finalFileName, "."+strings.ToLower(meta.Format))); err != nil {
			return fmt.Errorf("erro ao registrar formato: %w", err)
		}

		// 8. Inserir Sinopse em 'comments'
		if meta.Description != "" {
			_, _ = tx.Exec("INSERT INTO comments (book, text) VALUES (?, ?)", bookID, meta.Description)
		}

		// 9. Inserir Identificadores
		for k, v := range meta.Identifiers {
			_, _ = tx.Exec("INSERT INTO identifiers (book, type, val) VALUES (?, ?, ?)", bookID, k, v)
		}

		// 10. Atualizar caminho e flag de capa
		_, err = tx.Exec("UPDATE books SET path = ?, has_cover = ? WHERE id = ?", filepath.ToSlash(relPath), hasCover, bookID)
		if err != nil {
			return err
		}

		// 11. CWA Automatic EPUB Fixer Service
		if strings.EqualFold(meta.Format, "EPUB") {
			_, _ = formats.FixEPUB(finalFilePath)
		}

		// 12. CWA Automatic Metadata Fetch on Ingest (se capa ou sinopse estiverem ausentes)
		if hasCover == 0 || meta.Description == "" {
			go func(bID int64, bTitle, bAuthor, targetDir string, needCover, needDesc bool) {
				results := metadata.SearchOnline(context.Background(), bTitle+" "+bAuthor)
				if len(results) == 0 {
					return
				}
				first := results[0]
				_ = w.db.ExecuteWrite(func(txW *sql.Tx) error {
					if needDesc && first.Description != "" {
						_, _ = txW.Exec("INSERT OR REPLACE INTO comments (book, text) VALUES (?, ?)", bID, first.Description)
					}
					if needCover && first.Cover != "" {
						if resp, err := http.Get(first.Cover); err == nil && resp.StatusCode == 200 {
							cBytes, _ := io.ReadAll(resp.Body)
							resp.Body.Close()
							if len(cBytes) > 0 {
								cPath := filepath.Join(targetDir, "cover.jpg")
								if os.WriteFile(cPath, cBytes, 0644) == nil {
									_, _ = txW.Exec("UPDATE books SET has_cover = 1 WHERE id = ?", bID)
								}
							}
						}
					}
					return nil
				})
			}(bookID, title, author, absBookDir, hasCover == 0, meta.Description == "")
		}

		return nil
	})
}

func sanitizeName(n string) string {
	bad := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	res := n
	for _, b := range bad {
		res = strings.ReplaceAll(res, b, "_")
	}
	return strings.TrimSpace(res)
}

func moveOrCopy(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
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

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	in.Close()
	return os.Remove(src)
}
