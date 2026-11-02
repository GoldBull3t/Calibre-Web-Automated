package watcher

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/crocodilestick/calibre-web-go/internal/cleaner"
	"github.com/crocodilestick/calibre-web-go/internal/formats"
	"github.com/crocodilestick/calibre-web-go/internal/metadata"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

var coverClient = &http.Client{
	Timeout: 10 * time.Second,
}

func isSafeCoverURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	host := u.Hostname()
	if host == "localhost" || strings.EqualFold(host, "localhost.localdomain") {
		return false
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return false
		}
		if ip.String() == "169.254.169.254" {
			return false
		}
	}
	return true
}

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
	ingestDir      string
	libraryDir     string
	quarantineDir  string
	db             *database.DB
	timers         map[string]*time.Timer
	mu             sync.Mutex
	jobs           chan string
	converterPool  *formats.ConverterPool
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

	ext := strings.ToLower(filepath.Ext(srcPath))

	// 1. Unpacker Transiente (.zip e .rar): extrai e-books, descarta lixo promocional e agenda ingestão
	if ext == ".zip" || ext == ".rar" {
		extracted, unpackErr := cleaner.UnpackTransientArchive(srcPath, w.ingestDir)
		if unpackErr != nil {
			log.Printf("[UNPACK ERROR] Falha ao descompactar %s: %v", srcPath, unpackErr)
			_ = cleaner.QuarantineFile(context.Background(), srcPath, w.quarantineDir, "Falha na descompressão de arquivo", unpackErr, w.db)
			return nil
		}
		_ = os.Remove(srcPath)
		log.Printf("[UNPACK SUCCESS] %d e-book(s) extraídos de %s", len(extracted), filepath.Base(srcPath))
		for _, ebookFile := range extracted {
			w.debounce(ebookFile)
		}
		return nil
	}

	// 2. Content Deduplication via SHA-256
	fileHash, _, hashErr := cleaner.CalculateSHA256(srcPath)
	if hashErr == nil && fileHash != "" {
		if existingID, existingFmt, err := w.db.GetBookByHash(context.Background(), fileHash); err == nil && existingID > 0 {
			fmtClean := strings.ToUpper(strings.TrimPrefix(ext, "."))
			if strings.EqualFold(existingFmt, fmtClean) {
				log.Printf("[DEDUPLICATION] Ficheiro redundante descartado (SHA-256: %s): %s", fileHash[:12], filepath.Base(srcPath))
				_ = os.Remove(srcPath)
				return nil
			}
		}
	}

	// 3. Extração de Metadados com Mecanismo de Quarentena
	meta, err := formats.Extract(srcPath)
	if err != nil {
		log.Printf("[QUARANTINE] Metadados corrompidos ou ilegíveis em %s: %v. A mover para quarentena.", filepath.Base(srcPath), err)
		_ = cleaner.QuarantineFile(context.Background(), srcPath, w.quarantineDir, "Metadados corrompidos ou ilegíveis", err, w.db)
		return nil
	}

	// 4. Normalização Regex e Parser de Séries
	rawBase := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	cleanTitleStr, cleanAuthorStr := cleaner.CleanTitleAndAuthor(rawBase)

	title := meta.Title
	if title == "" || strings.EqualFold(title, "Unknown") {
		title = cleanTitleStr
	} else {
		title = cleaner.CleanString(title)
	}

	author := "Unknown"
	if len(meta.Authors) > 0 && strings.TrimSpace(meta.Authors[0]) != "" {
		author = cleaner.CleanString(meta.Authors[0])
	} else if cleanAuthorStr != "" {
		author = cleanAuthorStr
	}

	cleanTitle, seriesName, seriesIndex, hasSeries := cleaner.ParseSeries(rawBase)
	if !hasSeries && meta.Series != "" {
		seriesName = cleaner.CleanString(meta.Series)
		seriesIndex = meta.SeriesIndex
		hasSeries = true
	}
	if hasSeries && cleanTitle != "" {
		title = cleanTitle
	}
	if seriesIndex <= 0 {
		seriesIndex = 1.0
	}

	cleanAuthor := sanitizeName(author)
	cleanTitleClean := sanitizeName(title)

	// Otimização prévia de imagem de capa em memória (máx 1600px de altura)
	if len(meta.CoverBytes) > 0 {
		if optBytes, errO := formats.OptimizeCoverBytes(meta.CoverBytes, 1600); errO == nil && len(optBytes) > 0 {
			meta.CoverBytes = optBytes
		}
	}

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
				finalFileName := fmt.Sprintf("%s - %s.%s", cleanTitleClean, cleanAuthor, strings.ToLower(meta.Format))
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
				if fileHash != "" {
					_ = w.db.RegisterBookHash(context.Background(), fileHash, existingBookID, meta.Format, fi.Size())
				}
				return err
			}
			// Se o formato já existia, descarta o arquivo de ingestão duplicado
			_ = os.Remove(srcPath)
			return nil
		}

		// 3. Inserção de Novo Livro
		bookUUID := uuid.New().String()
		now := time.Now().UTC()

		res, err := tx.Exec(`
			INSERT INTO books (title, sort, timestamp, pubdate, series_index, last_modified, path, has_cover, uuid)
			VALUES (?, ?, ?, ?, ?, ?, '', 0, ?)
		`, title, title, now, now, seriesIndex, now, bookUUID)
		if err != nil {
			return fmt.Errorf("erro ao inserir livro: %w", err)
		}
		bookID, _ := res.LastInsertId()

		// 4. Vincular Livro ao Autor
		if _, err := tx.Exec("INSERT INTO books_authors_link (book, author) VALUES (?, ?)", bookID, authorID); err != nil {
			return fmt.Errorf("erro ao vincular autor: %w", err)
		}

		// 5. Vincular Série e Número de Índice se identificados
		if hasSeries && seriesName != "" {
			var seriesID int64
			sErr := tx.QueryRow("SELECT id FROM series WHERE name = ?", seriesName).Scan(&seriesID)
			if sErr == sql.ErrNoRows {
				resS, errS := tx.Exec("INSERT INTO series (name, sort) VALUES (?, ?)", seriesName, seriesName)
				if errS == nil {
					seriesID, _ = resS.LastInsertId()
				}
			}
			if seriesID > 0 {
				_, _ = tx.Exec("INSERT OR IGNORE INTO books_series_link (book, series) VALUES (?, ?)", bookID, seriesID)
			}
		}

		// 6. Estrutura de Diretórios Calibre: Autor/Título (ID)
		relPath := filepath.Join(cleanAuthor, fmt.Sprintf("%s (%d)", cleanTitleClean, bookID))
		absBookDir := filepath.Join(w.libraryDir, relPath)
		if err := os.MkdirAll(absBookDir, 0755); err != nil {
			return fmt.Errorf("erro ao criar pasta do livro: %w", err)
		}

		finalFileName := fmt.Sprintf("%s - %s.%s", cleanTitleClean, cleanAuthor, strings.ToLower(meta.Format))
		finalFilePath := filepath.Join(absBookDir, finalFileName)

		if err := moveOrCopy(srcPath, finalFilePath); err != nil {
			return fmt.Errorf("erro ao mover ficheiro do livro: %w", err)
		}

		// 7. Extração e Gravação de Capa
		hasCover := 0
		if len(meta.CoverBytes) > 0 {
			coverDest := filepath.Join(absBookDir, "cover.jpg")
			if err := os.WriteFile(coverDest, meta.CoverBytes, 0644); err == nil {
				hasCover = 1
			}
		}

		// 8. Inserir Formato em 'data'
		uncompressedSize := fi.Size()
		if _, err := tx.Exec(`
			INSERT INTO data (book, format, uncompressed_size, name)
			VALUES (?, ?, ?, ?)
		`, bookID, meta.Format, uncompressedSize, strings.TrimSuffix(finalFileName, "."+strings.ToLower(meta.Format))); err != nil {
			return fmt.Errorf("erro ao registrar formato: %w", err)
		}

		// Registo de Hash SHA-256 para desduplicação
		if fileHash != "" {
			_ = w.db.RegisterBookHash(context.Background(), fileHash, bookID, meta.Format, uncompressedSize)
		}

		// 9. Inserir Sinopse em 'comments'
		if meta.Description != "" {
			_, _ = tx.Exec("INSERT INTO comments (book, text) VALUES (?, ?)", bookID, meta.Description)
		}

		// 10. Inserir Identificadores
		for k, v := range meta.Identifiers {
			_, _ = tx.Exec("INSERT INTO identifiers (book, type, val) VALUES (?, ?, ?)", bookID, k, v)
		}

		// 11. Atualizar caminho e flag de capa
		_, err = tx.Exec("UPDATE books SET path = ?, has_cover = ? WHERE id = ?", filepath.ToSlash(relPath), hasCover, bookID)
		if err != nil {
			return err
		}

		// 12. Sanitização Cirúrgica de CSS/HTML e Geração de Kepub
		if strings.EqualFold(meta.Format, "EPUB") {
			_, _ = formats.SanitizeEPUB(finalFilePath, finalFilePath)

if w.converterPool != nil {
				w.converterPool.Submit(func(ctx context.Context) error {
					kepubPath, errK := formats.GenerateKepub(finalFilePath, absBookDir)
					if errK == nil && kepubPath != "" {
						fiK, errStat := os.Stat(kepubPath)
						if errStat == nil {
							_ = w.db.ExecuteWrite(func(txK *sql.Tx) error {
								_, errIns := txK.Exec(`
									INSERT INTO data (book, format, uncompressed_size, name)
									VALUES (?, 'KEPUB', ?, ?)
									ON CONFLICT(book, format) DO NOTHING
								`, bookID, fiK.Size(), strings.TrimSuffix(filepath.Base(kepubPath), ".kepub.epub"))
								return errIns
							})
						}
					}
					return nil
				})
			}
		}

		// 13. Metadados e capa automáticos online se ausentes
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
					if needCover && first.Cover != "" && isSafeCoverURL(first.Cover) {
						if resp, err := coverClient.Get(first.Cover); err == nil && resp.StatusCode == 200 {
							cBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
							resp.Body.Close()
							if len(cBytes) > 0 {
								if optC, errO := formats.OptimizeCoverBytes(cBytes, 1600); errO == nil {
									cBytes = optC
								}
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
