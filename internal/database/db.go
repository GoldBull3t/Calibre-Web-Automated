package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"modernc.org/sqlite"
)

func init() {
	// Regista funções SQL escalares esperadas pelos triggers do Calibre (ex: books_insert_trg, series_insert_trg)
	_ = sqlite.RegisterScalarFunction("title_sort", 1, func(ctx *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		if len(args) == 0 || args[0] == nil {
			return "", nil
		}
		s := fmt.Sprint(args[0])
		lower := strings.ToLower(s)
		for _, prefix := range []string{"the ", "a ", "an ", "o ", "os ", "a ", "as ", "um ", "uma "} {
			if strings.HasPrefix(lower, prefix) {
				return strings.TrimSpace(s[len(prefix):]) + ", " + strings.TrimSpace(s[:len(prefix)]), nil
			}
		}
		return strings.TrimSpace(s), nil
	})

	_ = sqlite.RegisterScalarFunction("uuid4", 0, func(ctx *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		return uuid.New().String(), nil
	})
}

type DB struct {
	roPool  *sql.DB // Pool para rotas de leitura com alta concorrência
	rwPool  *sql.DB // Conexão de escrita única para prevenir SQLite busy locks
	appPool *sql.DB // Conexão com app.db (utilizadores, estantes, progresso)
	mu      sync.RWMutex
}

// Open inicializa as ligações ao metadata.db e app.db com cache de páginas de 2MB por conexão e modo WAL.
func Open(metadataDbPath, appDbPath string) (*DB, error) {
	// 1. Conexão RW inicializada primeiro para assegurar journal_mode=WAL no cabeçalho do ficheiro
	rwDSN := fmt.Sprintf("file:%s?mode=rwc&_journal_mode=WAL&_cache_size=-2000&_busy_timeout=5000&_synchronous=NORMAL", metadataDbPath)
	rw, err := sql.Open("sqlite", rwDSN)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir DB em modo escrita: %w", err)
	}

	rw.SetMaxOpenConns(1)
	rw.SetMaxIdleConns(1)
	rw.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rw.ExecContext(ctx, `
		PRAGMA journal_mode=WAL;
		PRAGMA synchronous=NORMAL;
		PRAGMA mmap_size=67108864;
		PRAGMA temp_store=MEMORY;
		PRAGMA foreign_keys=ON;
	`); err != nil {
		rw.Close()
		return nil, fmt.Errorf("falha ao ativar WAL mode: %w", err)
	}

	// 2. Conexão RO aberta com WAL já consolidado e cache de 2MB
	roDSN := fmt.Sprintf("file:%s?mode=ro&_cache_size=-2000&_busy_timeout=5000", metadataDbPath)
	ro, err := sql.Open("sqlite", roDSN)
	if err != nil {
		rw.Close()
		return nil, fmt.Errorf("erro ao abrir DB em modo leitura: %w", err)
	}

	ro.SetMaxOpenConns(6)
	ro.SetMaxIdleConns(4)
	ro.SetConnMaxLifetime(30 * time.Minute)

	if err := ro.PingContext(ctx); err != nil {
		ro.Close()
		rw.Close()
		return nil, fmt.Errorf("ping de leitura falhou: %w", err)
	}

	// 3. Conexão opcional com app.db
	var app *sql.DB
	if appDbPath != "" {
		if _, err := os.Stat(appDbPath); err == nil {
			appDSN := fmt.Sprintf("file:%s?mode=rwc&_journal_mode=WAL&_cache_size=-2000&_busy_timeout=5000", appDbPath)
			if appConn, err := sql.Open("sqlite", appDSN); err == nil {
				appConn.SetMaxOpenConns(2)
				appConn.SetMaxIdleConns(2)
				app = appConn
			}
		}
	}

	return &DB{roPool: ro, rwPool: rw, appPool: app}, nil
}

// CheckpointWAL consolida os ficheiros WAL em disco e trunca o log para zero.
func (d *DB) CheckpointWAL(ctx context.Context) error {
	if d.rwPool != nil {
		_, _ = d.rwPool.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE);")
	}
	if d.appPool != nil {
		_, _ = d.appPool.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE);")
	}
	return nil
}

func (d *DB) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = d.CheckpointWAL(ctx)

	errRO := d.roPool.Close()
	errRW := d.rwPool.Close()
	if d.appPool != nil {
		_ = d.appPool.Close()
	}
	if errRO != nil {
		return errRO
	}
	return errRW
}

func (d *DB) AppDB() *sql.DB {
	return d.appPool
}

func (d *DB) MetadataDB() *sql.DB {
	return d.rwPool
}

// ExecuteWrite executa uma transação de escrita no metadata.db de forma atómica e thread-safe.
func (d *DB) ExecuteWrite(fn func(tx *sql.Tx) error) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.rwPool.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ListBooks executa query otimizada com paginação a nível de banco de dados e GROUP_CONCAT.
func (d *DB) ListBooks(ctx context.Context, f BookFilter) (*PageResult[Book], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 24
	}
	offset := (f.Page - 1) * f.PageSize

	var conditions []string
	var args []any

	if f.Search != "" {
		conditions = append(conditions, "(b.title LIKE ? OR a.name LIKE ? OR t.name LIKE ?)")
		args = append(args, "%"+f.Search+"%", "%"+f.Search+"%", "%"+f.Search+"%")
	}
	if f.AuthorID > 0 {
		conditions = append(conditions, "bal.author = ?")
		args = append(args, f.AuthorID)
	}
	if f.SeriesID > 0 {
		conditions = append(conditions, "bsl.series = ?")
		args = append(args, f.SeriesID)
	}
	if f.TagID > 0 {
		conditions = append(conditions, "btl.tag = ?")
		args = append(args, f.TagID)
	}
	if f.PublisherID > 0 {
		conditions = append(conditions, "bpl.publisher = ?")
		args = append(args, f.PublisherID)
	}
	if f.RatingID > 0 {
		conditions = append(conditions, "brl.rating = ?")
		args = append(args, f.RatingID)
	}
	if f.LanguageID > 0 {
		conditions = append(conditions, "bll.lang_code = ?")
		args = append(args, f.LanguageID)
	}
	if f.Format != "" {
		conditions = append(conditions, "d.format = ?")
		args = append(args, strings.ToUpper(f.Format))
	}
	if len(f.BookIDs) > 0 {
		placeholders := make([]string, len(f.BookIDs))
		for i, id := range f.BookIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		conditions = append(conditions, fmt.Sprintf("b.id IN (%s)", strings.Join(placeholders, ",")))
	}

	if f.ReadStatus > 0 && f.UserID > 0 && d.appPool != nil {
		var readIDs []int64
		rRows, rErr := d.appPool.QueryContext(ctx, "SELECT book_id FROM book_read_link WHERE user_id = ? AND read_status = 2", f.UserID)
		if rErr == nil {
			for rRows.Next() {
				var rid int64
				if rRows.Scan(&rid) == nil {
					readIDs = append(readIDs, rid)
				}
			}
			_ = rRows.Err()
			rRows.Close()
		}
		if f.ReadStatus == 1 { // Lido
			if len(readIDs) == 0 {
				return &PageResult[Book]{Items: []Book{}, Page: f.Page, PageSize: f.PageSize, TotalItems: 0, TotalPages: 0}, nil
			}
			placeholders := make([]string, len(readIDs))
			for i, rid := range readIDs {
				placeholders[i] = "?"
				args = append(args, rid)
			}
			conditions = append(conditions, fmt.Sprintf("b.id IN (%s)", strings.Join(placeholders, ",")))
		} else if f.ReadStatus == 2 { // Não lido
			if len(readIDs) > 0 {
				placeholders := make([]string, len(readIDs))
				for i, rid := range readIDs {
					placeholders[i] = "?"
					args = append(args, rid)
				}
				conditions = append(conditions, fmt.Sprintf("b.id NOT IN (%s)", strings.Join(placeholders, ",")))
			}
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderBy := "b.id DESC"
	switch f.Sort {
	case "old":
		orderBy = "b.id ASC"
	case "abc", "asc":
		orderBy = "b.sort COLLATE NOCASE ASC"
	case "zyx", "desc":
		orderBy = "b.sort COLLATE NOCASE DESC"
	case "authaz":
		orderBy = "authors_list COLLATE NOCASE ASC"
	case "authza":
		orderBy = "authors_list COLLATE NOCASE DESC"
	case "pubnew":
		orderBy = "b.pubdate DESC"
	case "pubold":
		orderBy = "b.pubdate ASC"
	case "random":
		orderBy = "RANDOM()"
	case "hotdesc":
		orderBy = "b.id DESC"
	default:
		orderBy = "b.id DESC"
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT b.id)
		FROM books b
		LEFT JOIN books_authors_link bal ON b.id = bal.book
		LEFT JOIN authors a ON bal.author = a.id
		LEFT JOIN books_series_link bsl ON b.id = bsl.book
		LEFT JOIN series s ON bsl.series = s.id
		LEFT JOIN books_tags_link btl ON b.id = btl.book
		LEFT JOIN tags t ON btl.tag = t.id
		LEFT JOIN books_publishers_link bpl ON b.id = bpl.book
		LEFT JOIN publishers p ON bpl.publisher = p.id
		LEFT JOIN books_ratings_link brl ON b.id = brl.book
		LEFT JOIN ratings r ON brl.rating = r.id
		LEFT JOIN books_languages_link bll ON b.id = bll.book
		LEFT JOIN languages l ON bll.lang_code = l.id
		LEFT JOIN data d ON b.id = d.book
		%s
	`, whereClause)

	var totalItems int64
	if err := d.roPool.QueryRowContext(ctx, countQuery, args...).Scan(&totalItems); err != nil {
		return nil, fmt.Errorf("falha ao contar livros: %w", err)
	}

	dataQuery := fmt.Sprintf(`
		SELECT 
			b.id, 
			b.title, 
			COALESCE(b.sort, ''), 
			b.path, 
			b.has_cover, 
			COALESCE(b.pubdate, CURRENT_TIMESTAMP), 
			b.series_index,
			COALESCE(GROUP_CONCAT(DISTINCT a.name), '') AS authors_list,
			COALESCE(s.name, '') AS series_name,
			COALESCE(GROUP_CONCAT(DISTINCT t.name), '') AS tags_list,
			COALESCE(GROUP_CONCAT(d.format || '::' || d.uncompressed_size || '::' || REPLACE(d.name, ',', '__COMMA__')), '') AS formats_list,
			COALESCE(GROUP_CONCAT(DISTINCT p.name), '') AS publishers_list,
			COALESCE(MAX(r.rating), 0) AS max_rating,
			COALESCE(GROUP_CONCAT(DISTINCT l.lang_code), '') AS languages_list
		FROM books b
		LEFT JOIN books_authors_link bal ON b.id = bal.book
		LEFT JOIN authors a ON bal.author = a.id
		LEFT JOIN books_series_link bsl ON b.id = bsl.book
		LEFT JOIN series s ON bsl.series = s.id
		LEFT JOIN books_tags_link btl ON b.id = btl.book
		LEFT JOIN tags t ON btl.tag = t.id
		LEFT JOIN books_publishers_link bpl ON b.id = bpl.book
		LEFT JOIN publishers p ON bpl.publisher = p.id
		LEFT JOIN books_ratings_link brl ON b.id = brl.book
		LEFT JOIN ratings r ON brl.rating = r.id
		LEFT JOIN books_languages_link bll ON b.id = bll.book
		LEFT JOIN languages l ON bll.lang_code = l.id
		LEFT JOIN data d ON b.id = d.book
		%s
		GROUP BY b.id
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, whereClause, orderBy)

	queryArgs := append(args, f.PageSize, offset)
	rows, err := d.roPool.QueryContext(ctx, dataQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar livros: %w", err)
	}
	defer rows.Close()

	books := make([]Book, 0, f.PageSize)
	for rows.Next() {
		var b Book
		var authorsStr, tagsStr, formatsStr, pubDateStr, publishersStr, languagesStr string

		err := rows.Scan(
			&b.ID,
			&b.Title,
			&b.Sort,
			&b.Path,
			&b.HasCover,
			&pubDateStr,
			&b.SeriesIndex,
			&authorsStr,
			&b.Series,
			&tagsStr,
			&formatsStr,
			&publishersStr,
			&b.Rating,
			&languagesStr,
		)
		if err != nil {
			return nil, fmt.Errorf("erro no scan do livro: %w", err)
		}

		b.PubDate = parseCalibreDate(pubDateStr)
		if authorsStr != "" {
			b.Authors = strings.Split(authorsStr, ",")
		}
		if tagsStr != "" {
			b.Tags = strings.Split(tagsStr, ",")
		}
		if formatsStr != "" {
			b.Formats = parseFormats(formatsStr)
		}
		if publishersStr != "" {
			b.Publishers = strings.Split(publishersStr, ",")
		}
		if languagesStr != "" {
			b.Languages = strings.Split(languagesStr, ",")
		}

		books = append(books, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar livros: %w", err)
	}

	if f.UserID > 0 && d.appPool != nil && len(books) > 0 {
		bIDs := make([]any, len(books))
		idIndex := make(map[int64]int, len(books))
		placeholders := make([]string, len(books))
		for i, b := range books {
			bIDs[i] = b.ID
			idIndex[b.ID] = i
			placeholders[i] = "?"
		}
		inClause := strings.Join(placeholders, ",")

		rQuery := fmt.Sprintf("SELECT book_id, read_status FROM book_read_link WHERE user_id = ? AND book_id IN (%s)", inClause)
		rRows, rErr := d.appPool.QueryContext(ctx, rQuery, append([]any{f.UserID}, bIDs...)...)
		if rErr == nil {
			for rRows.Next() {
				var bID int64
				var status int
				if rRows.Scan(&bID, &status) == nil {
					if idx, ok := idIndex[bID]; ok {
						books[idx].ReadStatus = (status == 2)
					}
				}
			}
			_ = rRows.Err()
			rRows.Close()
		}

		aQuery := fmt.Sprintf("SELECT book_id, is_archived FROM archived_book WHERE user_id = ? AND book_id IN (%s)", inClause)
		aRows, aErr := d.appPool.QueryContext(ctx, aQuery, append([]any{f.UserID}, bIDs...)...)
		if aErr == nil {
			for aRows.Next() {
				var bID int64
				var isArchived bool
				if aRows.Scan(&bID, &isArchived) == nil {
					if idx, ok := idIndex[bID]; ok {
						books[idx].IsArchived = isArchived
					}
				}
			}
			_ = aRows.Err()
			aRows.Close()
		}
	}

	totalPages := int((totalItems + int64(f.PageSize) - 1) / int64(f.PageSize))
	return &PageResult[Book]{
		Items:      books,
		TotalItems: totalItems,
		Page:       f.Page,
		PageSize:   f.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetBookByID retorna os detalhes do livro com sinopse e identificadores.
func (d *DB) GetBookByID(ctx context.Context, id int64) (*BookDetails, error) {
	query := `
		SELECT 
			b.id, b.title, COALESCE(b.sort, ''), b.path, b.has_cover, 
			COALESCE(b.pubdate, CURRENT_TIMESTAMP), b.series_index,
			COALESCE(GROUP_CONCAT(DISTINCT a.name), '') AS authors_list,
			COALESCE(s.name, '') AS series_name,
			COALESCE(GROUP_CONCAT(DISTINCT t.name), '') AS tags_list,
			COALESCE(GROUP_CONCAT(d.format || '::' || d.uncompressed_size || '::' || REPLACE(d.name, ',', '__COMMA__')), '') AS formats_list,
			COALESCE(GROUP_CONCAT(DISTINCT p.name), '') AS publishers_list,
			COALESCE(MAX(r.rating), 0) AS max_rating,
			COALESCE(GROUP_CONCAT(DISTINCT l.lang_code), '') AS languages_list,
			COALESCE(c.text, '') AS comment
		FROM books b
		LEFT JOIN books_authors_link bal ON b.id = bal.book
		LEFT JOIN authors a ON bal.author = a.id
		LEFT JOIN books_series_link bsl ON b.id = bsl.book
		LEFT JOIN series s ON bsl.series = s.id
		LEFT JOIN books_tags_link btl ON b.id = btl.book
		LEFT JOIN tags t ON btl.tag = t.id
		LEFT JOIN books_publishers_link bpl ON b.id = bpl.book
		LEFT JOIN publishers p ON bpl.publisher = p.id
		LEFT JOIN books_ratings_link brl ON b.id = brl.book
		LEFT JOIN ratings r ON brl.rating = r.id
		LEFT JOIN books_languages_link bll ON b.id = bll.book
		LEFT JOIN languages l ON bll.lang_code = l.id
		LEFT JOIN data d ON b.id = d.book
		LEFT JOIN comments c ON b.id = c.book
		WHERE b.id = ?
		GROUP BY b.id
	`
	var bd BookDetails
	var authorsStr, tagsStr, formatsStr, pubDateStr, publishersStr, languagesStr string

	err := d.roPool.QueryRowContext(ctx, query, id).Scan(
		&bd.ID, &bd.Title, &bd.Sort, &bd.Path, &bd.HasCover,
		&pubDateStr, &bd.SeriesIndex, &authorsStr, &bd.Series,
		&tagsStr, &formatsStr, &publishersStr, &bd.Rating, &languagesStr, &bd.Comment,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	bd.PubDate = parseCalibreDate(pubDateStr)
	if authorsStr != "" {
		bd.Authors = strings.Split(authorsStr, ",")
	}
	if tagsStr != "" {
		bd.Tags = strings.Split(tagsStr, ",")
	}
	if formatsStr != "" {
		bd.Formats = parseFormats(formatsStr)
	}

	// Consulta direta à tabela data para máxima resiliência de formatos e nomes de ficheiro
	formatRows, errFmt := d.roPool.QueryContext(ctx, "SELECT format, uncompressed_size, name FROM data WHERE book = ?", id)
	if errFmt == nil {
		defer formatRows.Close()
		var directFormats []DataFormat
		for formatRows.Next() {
			var f DataFormat
			if formatRows.Scan(&f.Format, &f.Size, &f.Name) == nil {
				directFormats = append(directFormats, f)
			}
		}
		_ = formatRows.Err()
		if len(directFormats) > 0 {
			bd.Formats = directFormats
		}
	}
	if publishersStr != "" {
		bd.Publishers = strings.Split(publishersStr, ",")
	}
	if languagesStr != "" {
		bd.Languages = strings.Split(languagesStr, ",")
	}

	bd.Identifiers = make(map[string]string)
	idRows, err := d.roPool.QueryContext(ctx, "SELECT type, val FROM identifiers WHERE book = ?", id)
	if err == nil {
		defer idRows.Close()
		for idRows.Next() {
			var k, v string
			if idRows.Scan(&k, &v) == nil {
				bd.Identifiers[k] = v
			}
		}
		_ = idRows.Err()
	}

	return &bd, nil
}

// GetStats retorna métricas rápidas da biblioteca para monitorização e painel.
func (d *DB) GetStats(ctx context.Context) (totalBooks int64, totalAuthors int64, totalSeries int64, err error) {
	_ = d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM books").Scan(&totalBooks)
	_ = d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM authors").Scan(&totalAuthors)
	_ = d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM series").Scan(&totalSeries)
	return totalBooks, totalAuthors, totalSeries, nil
}

// ListAuthors lista autores com contagem de livros e paginação.
func (d *DB) ListAuthors(ctx context.Context, page, pageSize int) (*PageResult[EntityCount], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 36
	}
	offset := (page - 1) * pageSize

	var totalItems int64
	if err := d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM authors").Scan(&totalItems); err != nil {
		return nil, err
	}

	query := `
		SELECT a.id, a.name, COUNT(bal.book) AS book_count
		FROM authors a
		LEFT JOIN books_authors_link bal ON a.id = bal.author
		GROUP BY a.id
		ORDER BY a.name COLLATE NOCASE ASC
		LIMIT ? OFFSET ?
	`
	rows, err := d.roPool.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]EntityCount, 0, pageSize)
	for rows.Next() {
		var ec EntityCount
		if err := rows.Scan(&ec.ID, &ec.Name, &ec.Count); err == nil {
			items = append(items, ec)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	return &PageResult[EntityCount]{
		Items:      items,
		TotalItems: totalItems,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ListSeries lista séries com contagem de livros e paginação.
func (d *DB) ListSeries(ctx context.Context, page, pageSize int) (*PageResult[EntityCount], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 36
	}
	offset := (page - 1) * pageSize

	var totalItems int64
	if err := d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM series").Scan(&totalItems); err != nil {
		return nil, err
	}

	query := `
		SELECT s.id, s.name, COUNT(bsl.book) AS book_count
		FROM series s
		LEFT JOIN books_series_link bsl ON s.id = bsl.series
		GROUP BY s.id
		ORDER BY s.name COLLATE NOCASE ASC
		LIMIT ? OFFSET ?
	`
	rows, err := d.roPool.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]EntityCount, 0, pageSize)
	for rows.Next() {
		var ec EntityCount
		if err := rows.Scan(&ec.ID, &ec.Name, &ec.Count); err == nil {
			items = append(items, ec)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	return &PageResult[EntityCount]{
		Items:      items,
		TotalItems: totalItems,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ListTags lista categorias com contagem de livros e paginação.
func (d *DB) ListTags(ctx context.Context, page, pageSize int) (*PageResult[EntityCount], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 36
	}
	offset := (page - 1) * pageSize

	var totalItems int64
	if err := d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM tags").Scan(&totalItems); err != nil {
		return nil, err
	}

	query := `
		SELECT t.id, t.name, COUNT(btl.book) AS book_count
		FROM tags t
		LEFT JOIN books_tags_link btl ON t.id = btl.tag
		GROUP BY t.id
		ORDER BY t.name COLLATE NOCASE ASC
		LIMIT ? OFFSET ?
	`
	rows, err := d.roPool.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]EntityCount, 0, pageSize)
	for rows.Next() {
		var ec EntityCount
		if err := rows.Scan(&ec.ID, &ec.Name, &ec.Count); err == nil {
			items = append(items, ec)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	return &PageResult[EntityCount]{
		Items:      items,
		TotalItems: totalItems,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ListPublishers lista editoras com contagem de livros e paginação.
func (d *DB) ListPublishers(ctx context.Context, page, pageSize int) (*PageResult[EntityCount], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 36
	}
	offset := (page - 1) * pageSize

	var totalItems int64
	if err := d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM publishers").Scan(&totalItems); err != nil {
		return nil, err
	}

	query := `
		SELECT p.id, p.name, COUNT(bpl.book) AS book_count
		FROM publishers p
		LEFT JOIN books_publishers_link bpl ON p.id = bpl.publisher
		GROUP BY p.id
		ORDER BY p.name COLLATE NOCASE ASC
		LIMIT ? OFFSET ?
	`
	rows, err := d.roPool.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]EntityCount, 0, pageSize)
	for rows.Next() {
		var ec EntityCount
		if err := rows.Scan(&ec.ID, &ec.Name, &ec.Count); err == nil {
			items = append(items, ec)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	return &PageResult[EntityCount]{
		Items:      items,
		TotalItems: totalItems,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ListRatings lista avaliações com contagem de livros e paginação.
func (d *DB) ListRatings(ctx context.Context, page, pageSize int) (*PageResult[EntityCount], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 36
	}
	offset := (page - 1) * pageSize

	var totalItems int64
	if err := d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM ratings").Scan(&totalItems); err != nil {
		return nil, err
	}

	query := `
		SELECT r.id, CAST(r.rating / 2 AS TEXT) || ' Stars' AS name, COUNT(brl.book) AS book_count
		FROM ratings r
		LEFT JOIN books_ratings_link brl ON r.id = brl.rating
		GROUP BY r.id
		ORDER BY r.rating DESC
		LIMIT ? OFFSET ?
	`
	rows, err := d.roPool.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]EntityCount, 0, pageSize)
	for rows.Next() {
		var ec EntityCount
		if err := rows.Scan(&ec.ID, &ec.Name, &ec.Count); err == nil {
			items = append(items, ec)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	return &PageResult[EntityCount]{
		Items:      items,
		TotalItems: totalItems,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ListLanguages lista línguas disponíveis com contagem de livros.
func (d *DB) ListLanguages(ctx context.Context, page, pageSize int) (*PageResult[EntityCount], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 36
	}
	offset := (page - 1) * pageSize

	var totalItems int64
	if err := d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM languages").Scan(&totalItems); err != nil {
		return nil, err
	}

	query := `
		SELECT l.id, l.lang_code AS name, COUNT(bll.book) AS book_count
		FROM languages l
		LEFT JOIN books_languages_link bll ON l.id = bll.lang_code
		GROUP BY l.id
		ORDER BY l.lang_code ASC
		LIMIT ? OFFSET ?
	`
	rows, err := d.roPool.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]EntityCount, 0, pageSize)
	for rows.Next() {
		var ec EntityCount
		if err := rows.Scan(&ec.ID, &ec.Name, &ec.Count); err == nil {
			items = append(items, ec)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	return &PageResult[EntityCount]{
		Items:      items,
		TotalItems: totalItems,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ListFormats lista todos os formatos disponíveis e quantidade de livros.
func (d *DB) ListFormats(ctx context.Context) ([]EntityCount, error) {
	query := `
		SELECT format, COUNT(DISTINCT book) AS book_count
		FROM data
		GROUP BY format
		ORDER BY format ASC
	`
	rows, err := d.roPool.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []EntityCount
	for rows.Next() {
		var format string
		var count int
		if err := rows.Scan(&format, &count); err == nil {
			items = append(items, EntityCount{
				Name:  format,
				Count: count,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetAuthorByID retorna o autor pelo ID.
func (d *DB) GetAuthorByID(ctx context.Context, id int64) (*EntityCount, error) {
	var ec EntityCount
	err := d.roPool.QueryRowContext(ctx, "SELECT id, name FROM authors WHERE id = ?", id).Scan(&ec.ID, &ec.Name)
	if err != nil {
		return nil, err
	}
	return &ec, nil
}

// GetSeriesByID retorna a série pelo ID.
func (d *DB) GetSeriesByID(ctx context.Context, id int64) (*EntityCount, error) {
	var ec EntityCount
	err := d.roPool.QueryRowContext(ctx, "SELECT id, name FROM series WHERE id = ?", id).Scan(&ec.ID, &ec.Name)
	if err != nil {
		return nil, err
	}
	return &ec, nil
}

// GetTagByID retorna a categoria/tag pelo ID.
func (d *DB) GetTagByID(ctx context.Context, id int64) (*EntityCount, error) {
	var ec EntityCount
	err := d.roPool.QueryRowContext(ctx, "SELECT id, name FROM tags WHERE id = ?", id).Scan(&ec.ID, &ec.Name)
	if err != nil {
		return nil, err
	}
	return &ec, nil
}

// GetPublisherByID retorna a editora pelo ID.
func (d *DB) GetPublisherByID(ctx context.Context, id int64) (*EntityCount, error) {
	var ec EntityCount
	err := d.roPool.QueryRowContext(ctx, "SELECT id, name FROM publishers WHERE id = ?", id).Scan(&ec.ID, &ec.Name)
	if err != nil {
		return nil, err
	}
	return &ec, nil
}

// GetAllAuthors retorna todos os autores ordenados alfabeticamente.
func (d *DB) GetAllAuthors(ctx context.Context) ([]EntityCount, error) {
	rows, err := d.roPool.QueryContext(ctx, "SELECT id, name FROM authors ORDER BY name COLLATE NOCASE ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []EntityCount
	for rows.Next() {
		var ec EntityCount
		if err := rows.Scan(&ec.ID, &ec.Name); err == nil {
			items = append(items, ec)
		}
	}
	return items, rows.Err()
}

// GetAllPublishers retorna todas as editoras.
func (d *DB) GetAllPublishers(ctx context.Context) ([]EntityCount, error) {
	rows, err := d.roPool.QueryContext(ctx, "SELECT id, name FROM publishers ORDER BY name COLLATE NOCASE ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []EntityCount
	for rows.Next() {
		var ec EntityCount
		if err := rows.Scan(&ec.ID, &ec.Name); err == nil {
			items = append(items, ec)
		}
	}
	return items, rows.Err()
}

// GetAllTags retorna todas as tags.
func (d *DB) GetAllTags(ctx context.Context) ([]EntityCount, error) {
	rows, err := d.roPool.QueryContext(ctx, "SELECT id, name FROM tags ORDER BY name COLLATE NOCASE ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []EntityCount
	for rows.Next() {
		var ec EntityCount
		if err := rows.Scan(&ec.ID, &ec.Name); err == nil {
			items = append(items, ec)
		}
	}
	return items, rows.Err()
}

// GetAllSeries retorna todas as séries.
func (d *DB) GetAllSeries(ctx context.Context) ([]EntityCount, error) {
	rows, err := d.roPool.QueryContext(ctx, "SELECT id, name FROM series ORDER BY name COLLATE NOCASE ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []EntityCount
	for rows.Next() {
		var ec EntityCount
		if err := rows.Scan(&ec.ID, &ec.Name); err == nil {
			items = append(items, ec)
		}
	}
	return items, rows.Err()
}

// GetAllLanguages retorna todas as linguagens.
func (d *DB) GetAllLanguages(ctx context.Context) ([]EntityCount, error) {
	rows, err := d.roPool.QueryContext(ctx, "SELECT id, lang_code FROM languages ORDER BY lang_code ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []EntityCount
	for rows.Next() {
		var ec EntityCount
		if err := rows.Scan(&ec.ID, &ec.Name); err == nil {
			items = append(items, ec)
		}
	}
	return items, rows.Err()
}

// DeleteBook remove um livro da biblioteca e limpa suas tabelas de junção.
func (d *DB) DeleteBook(ctx context.Context, id int64) error {
	return d.ExecuteWrite(func(tx *sql.Tx) error {
		tables := []string{
			"books_authors_link", "books_series_link", "books_tags_link",
			"books_publishers_link", "books_ratings_link", "books_languages_link",
			"data", "comments", "identifiers",
		}
		for _, t := range tables {
			_, _ = tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE book = ?", t), id)
		}
		_, err := tx.ExecContext(ctx, "DELETE FROM books WHERE id = ?", id)
		return err
	})
}

// GetUsers retorna a lista de utilizadores da base app.db.
func (d *DB) GetUsers(ctx context.Context) ([]User, error) {
	if d.appPool == nil {
		return []User{
			{ID: 1, Name: "admin", Email: "admin@example.org", Role: 479, IsAdmin: true, Locale: "en"},
		}, nil
	}

	rows, err := d.appPool.QueryContext(ctx, "SELECT id, name, email, role, COALESCE(kindle_mail, ''), COALESCE(locale, 'en') FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.KindleMail, &u.Locale); err == nil {
			u.IsAdmin = (u.Role&1) != 0 || (u.Role&16) != 0 || u.Role == 479
			users = append(users, u)
		}
	}
	return users, rows.Err()
}

// GetUserByName busca um utilizador pelo nome no app.db (case-insensitive).
func (d *DB) GetUserByName(ctx context.Context, name string) (*User, error) {
	if d.appPool == nil {
		if strings.EqualFold(name, "admin") {
			return &User{
				ID: 1, Name: "admin", Email: "admin@example.org", Role: 479, IsAdmin: true, Locale: "en",
			}, nil
		}
		return nil, sql.ErrNoRows
	}

	var u User
	err := d.appPool.QueryRowContext(ctx, `
		SELECT id, name, email, role, COALESCE(password, ''), COALESCE(kindle_mail, ''), COALESCE(locale, 'en'), COALESCE(default_language, 'en')
		FROM user
		WHERE name = ? COLLATE NOCASE
	`, name).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.PasswordHash, &u.KindleMail, &u.Locale, &u.DefaultLanguage)
	if err != nil {
		return nil, err
	}
	u.IsAdmin = (u.Role&1) != 0 || (u.Role&16) != 0 || u.Role == 479
	return &u, nil
}

// GetUserByEmail busca um utilizador pelo e-mail no app.db.
func (d *DB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if d.appPool == nil || strings.TrimSpace(email) == "" {
		return nil, sql.ErrNoRows
	}

	var u User
	err := d.appPool.QueryRowContext(ctx, `
		SELECT id, name, email, role, COALESCE(password, ''), COALESCE(kindle_mail, ''), COALESCE(locale, 'en'), COALESCE(default_language, 'en')
		FROM user
		WHERE email = ? COLLATE NOCASE
	`, strings.TrimSpace(email)).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.PasswordHash, &u.KindleMail, &u.Locale, &u.DefaultLanguage)
	if err != nil {
		return nil, err
	}
	u.IsAdmin = (u.Role&1) != 0 || (u.Role&16) != 0 || u.Role == 479
	return &u, nil
}

// GetUserByID busca um utilizador pelo ID.
func (d *DB) GetUserByID(ctx context.Context, id int64) (*User, error) {
	if d.appPool == nil {
		if id == 1 {
			return &User{
				ID: 1, Name: "admin", Email: "admin@example.org", Role: 479, IsAdmin: true, Locale: "en",
			}, nil
		}
		return nil, sql.ErrNoRows
	}

	var u User
	err := d.appPool.QueryRowContext(ctx, `
		SELECT id, name, email, role, COALESCE(password, ''), COALESCE(kindle_mail, ''), COALESCE(locale, 'en'), COALESCE(default_language, 'en')
		FROM user
		WHERE id = ?
	`, id).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.PasswordHash, &u.KindleMail, &u.Locale, &u.DefaultLanguage)
	if err != nil {
		return nil, err
	}
	u.IsAdmin = (u.Role&1) != 0 || (u.Role&16) != 0 || u.Role == 479
	return &u, nil
}

// CreateUserSession cria uma nova sessão no app.db para o utilizador.
func (d *DB) CreateUserSession(ctx context.Context, userID int64, sessionKey string) error {
	if d.appPool == nil {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, "INSERT INTO user_session (user_id, session_key) VALUES (?, ?)", userID, sessionKey)
	return err
}

// GetUserBySessionKey valida o sessionKey e devolve o utilizador associado.
func (d *DB) GetUserBySessionKey(ctx context.Context, sessionKey string) (*User, error) {
	if d.appPool == nil {
		if sessionKey != "" {
			return &User{ID: 1, Name: "admin", Email: "admin@example.org", Role: 479, IsAdmin: true, Locale: "en"}, nil
		}
		return nil, sql.ErrNoRows
	}

	var u User
	err := d.appPool.QueryRowContext(ctx, `
		SELECT u.id, u.name, u.email, u.role, COALESCE(u.kindle_mail, ''), COALESCE(u.locale, 'en'), COALESCE(u.default_language, 'en')
		FROM user u
		JOIN user_session s ON s.user_id = u.id
		WHERE s.session_key = ?
	`, sessionKey).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.KindleMail, &u.Locale, &u.DefaultLanguage)
	if err != nil {
		return nil, err
	}
	u.IsAdmin = (u.Role&1) != 0 || (u.Role&16) != 0 || u.Role == 479
	return &u, nil
}

// DeleteUserSession remove a sessão indicada.
func (d *DB) DeleteUserSession(ctx context.Context, sessionKey string) error {
	if d.appPool == nil {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, "DELETE FROM user_session WHERE session_key = ?", sessionKey)
	return err
}

// UpdateUserPassword atualiza o hash da senha do utilizador.
func (d *DB) UpdateUserPassword(ctx context.Context, userID int64, passwordHash string) error {
	if d.appPool == nil {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, "UPDATE user SET password = ? WHERE id = ?", passwordHash, userID)
	return err
}

// ToggleBookRead alterna o estado de leitura do livro para o utilizador no app.db.
func (d *DB) ToggleBookRead(ctx context.Context, userID, bookID int64) (bool, error) {
	if d.appPool == nil {
		return true, nil
	}

	var currentStatus int
	err := d.appPool.QueryRowContext(ctx, "SELECT read_status FROM book_read_link WHERE user_id = ? AND book_id = ?", userID, bookID).Scan(&currentStatus)
	now := time.Now().UTC()

	if err == sql.ErrNoRows {
		// Não existia: marca como lido (status = 2)
		_, err = d.appPool.ExecContext(ctx, `
			INSERT INTO book_read_link (book_id, user_id, read_status, last_modified, last_time_started_reading, times_started_reading)
			VALUES (?, ?, 2, ?, ?, 1)
		`, bookID, userID, now, now)
		return true, err
	} else if err != nil {
		return false, err
	}

	newStatus := 2
	if currentStatus == 2 {
		newStatus = 0
	}

	_, err = d.appPool.ExecContext(ctx, `
		UPDATE book_read_link SET read_status = ?, last_modified = ? WHERE user_id = ? AND book_id = ?
	`, newStatus, now, userID, bookID)
	return (newStatus == 2), err
}

// ToggleBookArchived alterna o estado de livro arquivado para o utilizador.
func (d *DB) ToggleBookArchived(ctx context.Context, userID, bookID int64) (bool, error) {
	if d.appPool == nil {
		return false, nil
	}

	var isArchived bool
	err := d.appPool.QueryRowContext(ctx, "SELECT is_archived FROM archived_book WHERE user_id = ? AND book_id = ?", userID, bookID).Scan(&isArchived)
	now := time.Now().UTC()

	if err == sql.ErrNoRows {
		_, err = d.appPool.ExecContext(ctx, `
			INSERT INTO archived_book (user_id, book_id, is_archived, last_modified)
			VALUES (?, ?, 1, ?)
		`, userID, bookID, now)
		return true, err
	} else if err != nil {
		return false, err
	}

	newStatus := !isArchived
	_, err = d.appPool.ExecContext(ctx, `
		UPDATE archived_book SET is_archived = ?, last_modified = ? WHERE user_id = ? AND book_id = ?
	`, newStatus, now, userID, bookID)
	return newStatus, err
}

// SaveBookmark guarda a posição de leitura do utilizador.
func (d *DB) SaveBookmark(ctx context.Context, userID, bookID int64, format, bookmarkKey string) error {
	if d.appPool == nil {
		return nil
	}
	_, _ = d.appPool.ExecContext(ctx, "DELETE FROM bookmark WHERE user_id = ? AND book_id = ? AND format = ? COLLATE NOCASE", userID, bookID, format)
	if bookmarkKey == "" {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, `
		INSERT INTO bookmark (user_id, book_id, format, bookmark_key)
		VALUES (?, ?, ?, ?)
	`, userID, bookID, strings.ToUpper(format), bookmarkKey)
	return err
}

// GetBookmark recupera a posição de leitura salva.
func (d *DB) GetBookmark(ctx context.Context, userID, bookID int64, format string) (string, error) {
	if d.appPool == nil {
		return "", nil
	}
	var bookmarkKey string
	err := d.appPool.QueryRowContext(ctx, "SELECT bookmark_key FROM bookmark WHERE user_id = ? AND book_id = ? AND format = ? COLLATE NOCASE", userID, bookID, format).Scan(&bookmarkKey)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return bookmarkKey, err
}

// GetBookUserState retorna (isRead, isArchived) para o utilizador.
func (d *DB) GetBookUserState(ctx context.Context, userID, bookID int64) (bool, bool, error) {
	if d.appPool == nil || userID <= 0 {
		return false, false, nil
	}

	var status int
	_ = d.appPool.QueryRowContext(ctx, "SELECT read_status FROM book_read_link WHERE user_id = ? AND book_id = ?", userID, bookID).Scan(&status)

	var isArchived bool
	_ = d.appPool.QueryRowContext(ctx, "SELECT is_archived FROM archived_book WHERE user_id = ? AND book_id = ?", userID, bookID).Scan(&isArchived)

	return (status == 2), isArchived, nil
}

// UpdateBookMetadata atualiza os metadados de um livro de forma atómica.
func (d *DB) UpdateBookMetadata(ctx context.Context, id int64, form BookEditForm) error {
	return d.ExecuteWrite(func(tx *sql.Tx) error {
		// 1. Atualiza dados principais em books
		titleSort := form.TitleSort
		if titleSort == "" {
			titleSort = form.Title
		}
		_, err := tx.ExecContext(ctx, "UPDATE books SET title = ?, sort = ?, series_index = ? WHERE id = ?",
			form.Title, titleSort, form.SeriesIndex, id)
		if err != nil {
			return err
		}

		// 2. Atualiza sinopse em comments
		_, _ = tx.ExecContext(ctx, "DELETE FROM comments WHERE book = ?", id)
		if form.Comment != "" {
			_, _ = tx.ExecContext(ctx, "INSERT INTO comments (book, text) VALUES (?, ?)", id, form.Comment)
		}

		// 3. Atualiza autores
		_, _ = tx.ExecContext(ctx, "DELETE FROM books_authors_link WHERE book = ?", id)
		for _, a := range form.Authors {
			a = strings.TrimSpace(a)
			if a == "" {
				continue
			}
			var authorID int64
			err := tx.QueryRowContext(ctx, "SELECT id FROM authors WHERE name = ?", a).Scan(&authorID)
			if err == sql.ErrNoRows {
				res, err := tx.ExecContext(ctx, "INSERT INTO authors (name, sort) VALUES (?, ?)", a, a)
				if err == nil {
					authorID, _ = res.LastInsertId()
				}
			}
			if authorID > 0 {
				_, _ = tx.ExecContext(ctx, "INSERT INTO books_authors_link (book, author) VALUES (?, ?)", id, authorID)
			}
		}

		// 4. Atualiza séries
		_, _ = tx.ExecContext(ctx, "DELETE FROM books_series_link WHERE book = ?", id)
		if s := strings.TrimSpace(form.Series); s != "" {
			var seriesID int64
			err := tx.QueryRowContext(ctx, "SELECT id FROM series WHERE name = ?", s).Scan(&seriesID)
			if err == sql.ErrNoRows {
				res, err := tx.ExecContext(ctx, "INSERT INTO series (name, sort) VALUES (?, ?)", s, s)
				if err == nil {
					seriesID, _ = res.LastInsertId()
				}
			}
			if seriesID > 0 {
				_, _ = tx.ExecContext(ctx, "INSERT INTO books_series_link (book, series) VALUES (?, ?)", id, seriesID)
			}
		}

		// 5. Atualiza tags
		_, _ = tx.ExecContext(ctx, "DELETE FROM books_tags_link WHERE book = ?", id)
		for _, t := range form.Tags {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			var tagID int64
			err := tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = ?", t).Scan(&tagID)
			if err == sql.ErrNoRows {
				res, err := tx.ExecContext(ctx, "INSERT INTO tags (name) VALUES (?)", t)
				if err == nil {
					tagID, _ = res.LastInsertId()
				}
			}
			if tagID > 0 {
				_, _ = tx.ExecContext(ctx, "INSERT INTO books_tags_link (book, tag) VALUES (?, ?)", id, tagID)
			}
		}

		// 6. Atualiza editoras
		_, _ = tx.ExecContext(ctx, "DELETE FROM books_publishers_link WHERE book = ?", id)
		for _, p := range form.Publishers {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			var pubID int64
			err := tx.QueryRowContext(ctx, "SELECT id FROM publishers WHERE name = ?", p).Scan(&pubID)
			if err == sql.ErrNoRows {
				res, err := tx.ExecContext(ctx, "INSERT INTO publishers (name, sort) VALUES (?, ?)", p, p)
				if err == nil {
					pubID, _ = res.LastInsertId()
				}
			}
			if pubID > 0 {
				_, _ = tx.ExecContext(ctx, "INSERT INTO books_publishers_link (book, publisher) VALUES (?, ?)", id, pubID)
			}
		}

		// 7. Atualiza ratings
		_, _ = tx.ExecContext(ctx, "DELETE FROM books_ratings_link WHERE book = ?", id)
		if form.Rating > 0 {
			var rID int64
			err := tx.QueryRowContext(ctx, "SELECT id FROM ratings WHERE rating = ?", form.Rating).Scan(&rID)
			if err == sql.ErrNoRows {
				res, err := tx.ExecContext(ctx, "INSERT INTO ratings (rating) VALUES (?)", form.Rating)
				if err == nil {
					rID, _ = res.LastInsertId()
				}
			}
			if rID > 0 {
				_, _ = tx.ExecContext(ctx, "INSERT INTO books_ratings_link (book, rating) VALUES (?, ?)", id, rID)
			}
		}

		// 8. Atualiza identificadores
		_, _ = tx.ExecContext(ctx, "DELETE FROM identifiers WHERE book = ?", id)
		for k, v := range form.Identifiers {
			k, v = strings.TrimSpace(k), strings.TrimSpace(v)
			if k != "" && v != "" {
				_, _ = tx.ExecContext(ctx, "INSERT INTO identifiers (book, type, val) VALUES (?, ?, ?)", id, k, v)
			}
		}

		return nil
	})
}

// DeleteBookFormat remove um ficheiro de formato associado a um livro.
func (d *DB) DeleteBookFormat(ctx context.Context, bookID int64, format string) error {
	return d.ExecuteWrite(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "DELETE FROM data WHERE book = ? AND LOWER(format) = LOWER(?)", bookID, format)
		return err
	})
}

// SetBookHasCover atualiza o sinalizador has_cover na tabela books.
func (d *DB) SetBookHasCover(ctx context.Context, id int64, hasCover bool) error {
	val := 0
	if hasCover {
		val = 1
	}
	return d.ExecuteWrite(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "UPDATE books SET has_cover = ? WHERE id = ?", val, id)
		return err
	})
}

// SetBookIdentifier insere ou atualiza um identificador (ex: hardcover, isbn, google) para um livro.
func (d *DB) SetBookIdentifier(ctx context.Context, bookID int64, idType, idVal string) error {
	idType = strings.TrimSpace(strings.ToLower(idType))
	idVal = strings.TrimSpace(idVal)
	if idType == "" || idVal == "" {
		return nil
	}
	return d.ExecuteWrite(func(tx *sql.Tx) error {
		_, _ = tx.ExecContext(ctx, "DELETE FROM identifiers WHERE book = ? AND LOWER(type) = ?", bookID, idType)
		_, err := tx.ExecContext(ctx, "INSERT INTO identifiers (book, type, val) VALUES (?, ?, ?)", bookID, idType, idVal)
		return err
	})
}

// GetBooksMissingIdentifier retorna livros que ainda não possuem um determinado identificador (ex: hardcover).
func (d *DB) GetBooksMissingIdentifier(ctx context.Context, idType string, limit int) ([]Book, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT b.id, b.title, b.path, b.has_cover,
		       COALESCE((SELECT GROUP_CONCAT(a.name, ' & ') FROM authors a JOIN books_authors_link bal ON a.id = bal.author WHERE bal.book = b.id), 'Autor Desconhecido') as authors
		FROM books b
		WHERE NOT EXISTS (
			SELECT 1 FROM identifiers i WHERE i.book = b.id AND LOWER(i.type) = ?
		)
		ORDER BY b.id DESC
		LIMIT ?`
	rows, err := d.roPool.QueryContext(ctx, query, strings.ToLower(idType), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Book
	for rows.Next() {
		var item Book
		var authorsStr string
		if err := rows.Scan(&item.ID, &item.Title, &item.Path, &item.HasCover, &authorsStr); err == nil {
			if authorsStr != "" {
				item.Authors = strings.Split(authorsStr, " & ")
			}
			list = append(list, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// AddBookFormat insere ou atualiza um formato na tabela data para o livro.
func (d *DB) AddBookFormat(ctx context.Context, bookID int64, format string, size int64, name string) error {
	format = strings.ToUpper(strings.TrimPrefix(format, "."))
	return d.ExecuteWrite(func(tx *sql.Tx) error {
		var exists int
		_ = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM data WHERE book = ? AND format = ?", bookID, format).Scan(&exists)
		if exists > 0 {
			_, err := tx.ExecContext(ctx, `
				UPDATE data SET uncompressed_size = ?, name = ? WHERE book = ? AND format = ?
			`, size, name, bookID, format)
			return err
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO data (book, format, uncompressed_size, name)
			VALUES (?, ?, ?, ?)
		`, bookID, format, size, name)
		return err
	})
}

// FindDuplicates deteta livros duplicados com base em títulos normalizados e autores.
func (d *DB) FindDuplicates(ctx context.Context) ([]DuplicateGroup, error) {
	query := `
		SELECT 
			b.id, b.title, COALESCE(GROUP_CONCAT(DISTINCT a.name), '') AS authors_list
		FROM books b
		LEFT JOIN books_authors_link bal ON b.id = bal.book
		LEFT JOIN authors a ON bal.author = a.id
		GROUP BY b.id
		ORDER BY b.id ASC
	`
	rows, err := d.roPool.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type rawBook struct {
		id      int64
		title   string
		authors string
		normKey string
	}

	normalize := func(s string) string {
		s = strings.ToLower(s)
		for _, prefix := range []string{"the ", "a ", "an ", "o ", "a ", "os ", "as "} {
			if strings.HasPrefix(s, prefix) {
				s = s[len(prefix):]
				break
			}
		}
		var clean strings.Builder
		for _, r := range s {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				clean.WriteRune(r)
			}
		}
		return clean.String()
	}

	groupsMap := make(map[string][]rawBook)
	for rows.Next() {
		var rb rawBook
		if err := rows.Scan(&rb.id, &rb.title, &rb.authors); err == nil {
			rb.normKey = normalize(rb.title)
			if len(rb.normKey) > 3 {
				groupsMap[rb.normKey] = append(groupsMap[rb.normKey], rb)
			}
		}
	}
	_ = rows.Err()

	var duplicateGroups []DuplicateGroup
	dismissed, _ := d.GetDismissedDuplicateHashes(ctx, 1)
	for key, group := range groupsMap {
		if len(group) > 1 {
			hash := fmt.Sprintf("grp_%s", key)
			if dismissed[hash] {
				continue
			}
			dg := DuplicateGroup{
				Hash:       hash,
				Title:      group[0].title,
				Confidence: 95,
			}
			for _, item := range group {
				b, _ := d.GetBookByID(ctx, item.id)
				if b != nil {
					dg.Books = append(dg.Books, *b)
				}
			}
			if len(dg.Books) > 1 {
				duplicateGroups = append(duplicateGroups, dg)
			}
		}
	}

	return duplicateGroups, nil
}

// EnsureDismissedDuplicatesTable garante a tabela de duplicados descartados no app.db.
func (d *DB) EnsureDismissedDuplicatesTable(ctx context.Context) error {
	if d.appPool == nil {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS dismissed_duplicate_groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			group_hash VARCHAR(64) NOT NULL,
			dismissed_at DATETIME,
			UNIQUE(user_id, group_hash)
		);
		CREATE INDEX IF NOT EXISTS idx_dismissed_dups ON dismissed_duplicate_groups(user_id, group_hash);
	`)
	return err
}

// DismissDuplicateGroup descarta um grupo de duplicados para o utilizador.
func (d *DB) DismissDuplicateGroup(ctx context.Context, userID int64, groupHash string) error {
	if d.appPool == nil {
		return nil
	}
	_ = d.EnsureDismissedDuplicatesTable(ctx)
	_, err := d.appPool.ExecContext(ctx, `
		INSERT OR IGNORE INTO dismissed_duplicate_groups (user_id, group_hash, dismissed_at)
		VALUES (?, ?, datetime('now'))
	`, userID, groupHash)
	return err
}

// UndismissDuplicateGroup reativa um grupo de duplicados descartado.
func (d *DB) UndismissDuplicateGroup(ctx context.Context, userID int64, groupHash string) error {
	if d.appPool == nil {
		return nil
	}
	_ = d.EnsureDismissedDuplicatesTable(ctx)
	_, err := d.appPool.ExecContext(ctx, `
		DELETE FROM dismissed_duplicate_groups WHERE user_id = ? AND group_hash = ?
	`, userID, groupHash)
	return err
}

// GetDismissedDuplicateHashes retorna o conjunto de hashes descartados para o utilizador.
func (d *DB) GetDismissedDuplicateHashes(ctx context.Context, userID int64) (map[string]bool, error) {
	dismissed := make(map[string]bool)
	if d.appPool == nil {
		return dismissed, nil
	}
	_ = d.EnsureDismissedDuplicatesTable(ctx)
	rows, err := d.appPool.QueryContext(ctx, "SELECT group_hash FROM dismissed_duplicate_groups WHERE user_id = ?", userID)
	if err != nil {
		return dismissed, err
	}
	defer rows.Close()
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err == nil {
			dismissed[h] = true
		}
	}
	return dismissed, rows.Err()
}

// GetLibraryStats calcula métricas reais da biblioteca.
func (d *DB) GetLibraryStats(ctx context.Context) (map[string]any, error) {
	var totalBooks, totalAuthors, totalSeries, totalTags, totalFormats, totalSize int64

	_ = d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM books").Scan(&totalBooks)
	_ = d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM authors").Scan(&totalAuthors)
	_ = d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM series").Scan(&totalSeries)
	_ = d.roPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM tags").Scan(&totalTags)
	_ = d.roPool.QueryRowContext(ctx, "SELECT COUNT(*), COALESCE(SUM(uncompressed_size), 0) FROM data").Scan(&totalFormats, &totalSize)

	return map[string]any{
		"total_books":   totalBooks,
		"total_authors": totalAuthors,
		"total_series":  totalSeries,
		"total_tags":    totalTags,
		"total_formats": totalFormats,
		"total_size":    totalSize,
	}, nil
}


// CreateShelf cria uma nova estante no app.db.
func (d *DB) CreateShelf(ctx context.Context, userID int64, name string, isPublic bool) (int64, error) {
	if d.appPool == nil {
		return 1, nil
	}
	pub := 0
	if isPublic {
		pub = 1
	}
	res, err := d.appPool.ExecContext(ctx, "INSERT INTO shelf (uuid, name, is_public, user_id, created, last_modified) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		uuid.New().String(), name, pub, userID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateShelf atualiza as propriedades de uma estante.
func (d *DB) UpdateShelf(ctx context.Context, shelfID int64, name string, isPublic bool) error {
	if d.appPool == nil {
		return nil
	}
	pub := 0
	if isPublic {
		pub = 1
	}
	_, err := d.appPool.ExecContext(ctx, "UPDATE shelf SET name = ?, is_public = ?, last_modified = CURRENT_TIMESTAMP WHERE id = ?",
		name, pub, shelfID)
	return err
}

// DeleteShelf remove uma estante e seus links do app.db.
func (d *DB) DeleteShelf(ctx context.Context, shelfID int64) error {
	if d.appPool == nil {
		return nil
	}
	_, _ = d.appPool.ExecContext(ctx, "DELETE FROM book_shelf_link WHERE shelf = ?", shelfID)
	_, err := d.appPool.ExecContext(ctx, "DELETE FROM shelf WHERE id = ?", shelfID)
	return err
}

// AddBookToShelf associa um livro à estante.
func (d *DB) AddBookToShelf(ctx context.Context, shelfID, bookID int64) error {
	if d.appPool == nil {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, "INSERT OR IGNORE INTO book_shelf_link (book_id, shelf, date_added) VALUES (?, ?, CURRENT_TIMESTAMP)",
		bookID, shelfID)
	return err
}

// RemoveBookFromShelf desassocia um livro da estante.
func (d *DB) RemoveBookFromShelf(ctx context.Context, shelfID, bookID int64) error {
	if d.appPool == nil {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, "DELETE FROM book_shelf_link WHERE shelf = ? AND book_id = ?",
		shelfID, bookID)
	return err
}

// GetShelfByID obtém a estante pelo ID.
func (d *DB) GetShelfByID(ctx context.Context, shelfID int64) (*Shelf, error) {
	if d.appPool == nil {
		return &Shelf{ID: shelfID, Name: fmt.Sprintf("Shelf %d", shelfID), IsPublic: true}, nil
	}
	var s Shelf
	var isPub int
	err := d.appPool.QueryRowContext(ctx, "SELECT id, name, is_public FROM shelf WHERE id = ?", shelfID).Scan(&s.ID, &s.Name, &isPub)
	if err != nil {
		return nil, err
	}
	s.IsPublic = isPub == 1
	return &s, nil
}

// ListCustomShelves devolve as estantes do utilizador ou públicas com contagem de livros.
func (d *DB) ListCustomShelves(ctx context.Context, userID int64) ([]Shelf, error) {
	if d.appPool == nil {
		return nil, nil
	}
	rows, err := d.appPool.QueryContext(ctx, `
		SELECT s.id, s.name, s.is_public, COUNT(bsl.book_id) as book_count
		FROM shelf s
		LEFT JOIN book_shelf_link bsl ON s.id = bsl.shelf
		WHERE s.is_public = 1 OR s.user_id = ?
		GROUP BY s.id
		ORDER BY s.name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Shelf
	for rows.Next() {
		var s Shelf
		var isPub int
		if err := rows.Scan(&s.ID, &s.Name, &isPub, &s.BookCount); err == nil {
			s.IsPublic = isPub == 1
			list = append(list, s)
		}
	}
	return list, rows.Err()
}

// GetShelfBooks obtém os livros de uma estante.
func (d *DB) GetShelfBooks(ctx context.Context, shelfID int64, sort string, page, pageSize int) (*PageResult[Book], error) {
	if d.appPool == nil {
		return d.ListBooks(ctx, BookFilter{Page: page, PageSize: pageSize, Sort: sort})
	}

	rows, err := d.appPool.QueryContext(ctx, "SELECT book_id FROM book_shelf_link WHERE shelf = ?", shelfID)
	if err != nil {
		return d.ListBooks(ctx, BookFilter{Page: page, PageSize: pageSize, Sort: sort})
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	_ = rows.Err()

	if len(ids) == 0 {
		return &PageResult[Book]{Items: []Book{}, TotalItems: 0, Page: page, PageSize: pageSize, TotalPages: 0}, nil
	}

	return d.ListBooks(ctx, BookFilter{
		BookIDs:  ids,
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Search:   "",
	})
}

// EnsureMagicShelfTable garante a tabela magic_shelf e os templates padrão de sistema.
func (d *DB) EnsureMagicShelfTable(ctx context.Context) error {
	if d.appPool == nil {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS magic_shelf (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			icon VARCHAR(32),
			rules TEXT,
			is_public BOOLEAN DEFAULT 1,
			user_id INTEGER NOT NULL DEFAULT 1,
			is_system BOOLEAN DEFAULT 1,
			created DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_modified DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return err
	}

	var count int
	_ = d.appPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM magic_shelf").Scan(&count)
	if count == 0 {
		_, _ = d.appPool.ExecContext(ctx, `
			INSERT INTO magic_shelf (name, description, icon, rules, is_public, user_id, is_system)
			VALUES 
			('Recently Added', 'Books added in the last 30 days', 'glyphicon-time', '{"field":"timestamp","operator":"newer_than","value":30}', 1, 1, 1),
			('Highly Rated', 'Books with 4 or 5 stars rating', 'glyphicon-star', '{"field":"rating","operator":"greater_or_equal","value":8}', 1, 1, 1),
			('Currently Reading', 'Books currently being read via KOReader or web', 'glyphicon-bookmark', '{"field":"read_status","operator":"reading"}', 1, 1, 1),
			('Books Without Covers', 'Books missing cover image', 'glyphicon-picture', '{"field":"has_cover","operator":"equal","value":0}', 1, 1, 1);
		`)
	}
	return nil
}

// ListMagicShelves retorna todas as estantes dinâmicas visíveis.
func (d *DB) ListMagicShelves(ctx context.Context, userID int64) ([]MagicShelf, error) {
	if d.appPool == nil {
		return nil, nil
	}
	_ = d.EnsureMagicShelfTable(ctx)

	rows, err := d.appPool.QueryContext(ctx, `
		SELECT id, name, COALESCE(description, ''), COALESCE(icon, 'glyphicon-book'), COALESCE(rules, ''), is_public, user_id
		FROM magic_shelf
		WHERE is_public = 1 OR user_id = ?
		ORDER BY id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []MagicShelf
	for rows.Next() {
		var ms MagicShelf
		var pub int
		if err := rows.Scan(&ms.ID, &ms.Name, &ms.Description, &ms.Icon, &ms.Rules, &pub, &ms.UserID); err == nil {
			ms.IsPublic = (pub == 1)
			list = append(list, ms)
		}
	}
	return list, rows.Err()
}

// GetMagicShelfBooks avalia as regras da estante dinâmica e retorna os livros correspondentes.
func (d *DB) GetMagicShelfBooks(ctx context.Context, shelfID int64, userID int64, page, pageSize int) (*PageResult[Book], error) {
	if d.appPool == nil {
		return d.ListBooks(ctx, BookFilter{Page: page, PageSize: pageSize})
	}
	_ = d.EnsureMagicShelfTable(ctx)

	var name, rules string
	err := d.appPool.QueryRowContext(ctx, "SELECT name, COALESCE(rules, '') FROM magic_shelf WHERE id = ?", shelfID).Scan(&name, &rules)
	if err != nil {
		return nil, err
	}

	filter := BookFilter{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
	}

	if strings.Contains(name, "Recently") || strings.Contains(rules, "newer_than") {
		filter.Sort = "new"
	} else if strings.Contains(name, "Highly") || strings.Contains(rules, "rating") {
		filter.RatingID = 4 // 4 ou 5 estrelas
	} else if strings.Contains(name, "Reading") || strings.Contains(rules, "reading") {
		rows, err := d.appPool.QueryContext(ctx, `
			SELECT DISTINCT CAST(document AS INTEGER) FROM kosync_progress WHERE user_id = ? AND percentage > 0 AND percentage < 99
			UNION
			SELECT DISTINCT book FROM bookmark WHERE user_id = ?
		`, userID, userID)
		if err == nil {
			defer rows.Close()
			var bookIDs []int64
			for rows.Next() {
				var bid int64
				if err := rows.Scan(&bid); err == nil && bid > 0 {
					bookIDs = append(bookIDs, bid)
				}
			}
			_ = rows.Err()
			filter.BookIDs = bookIDs
		}
	}

	return d.ListBooks(ctx, filter)
}

// CreateUser cria um novo utilizador no app.db.
func (d *DB) CreateUser(ctx context.Context, name, email, password string, role int) (int64, error) {
	if d.appPool == nil {
		return 1, nil
	}
	res, err := d.appPool.ExecContext(ctx, "INSERT INTO user (name, email, password, role, locale) VALUES (?, ?, ?, ?, 'en')",
		name, email, password, role)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateUser atualiza um utilizador no app.db.
func (d *DB) UpdateUser(ctx context.Context, userID int64, email, kindleMail string, role int) error {
	if d.appPool == nil {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, "UPDATE user SET email = ?, kindle_mail = ?, role = ? WHERE id = ?",
		email, kindleMail, role, userID)
	return err
}

// UpdateUserProfile atualiza as preferências e credenciais do utilizador.
func (d *DB) UpdateUserProfile(ctx context.Context, userID int64, email, kindleMail, locale, defaultLang, passwordHash string) error {
	if d.appPool == nil {
		return nil
	}

	if passwordHash != "" {
		_, err := d.appPool.ExecContext(ctx, `
			UPDATE user 
			SET email = ?, kindle_mail = ?, locale = ?, default_language = ?, password = ?
			WHERE id = ?
		`, email, kindleMail, locale, defaultLang, passwordHash, userID)
		return err
	}

	_, err := d.appPool.ExecContext(ctx, `
		UPDATE user 
		SET email = ?, kindle_mail = ?, locale = ?, default_language = ?
		WHERE id = ?
	`, email, kindleMail, locale, defaultLang, userID)
	return err
}

// BulkSetReadStatus atualiza em lote o estado de leitura dos livros para o utilizador.
func (d *DB) BulkSetReadStatus(ctx context.Context, userID int64, bookIDs []int64, markAsRead bool) error {
	if d.appPool == nil || len(bookIDs) == 0 {
		return nil
	}
	now := time.Now().UTC()
	status := 0
	if markAsRead {
		status = 2
	}
	for _, bID := range bookIDs {
		var exists int
		err := d.appPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM book_read_link WHERE user_id = ? AND book_id = ?", userID, bID).Scan(&exists)
		if err == nil && exists > 0 {
			_, _ = d.appPool.ExecContext(ctx, "UPDATE book_read_link SET read_status = ?, last_modified = ? WHERE user_id = ? AND book_id = ?", status, now, userID, bID)
		} else {
			_, _ = d.appPool.ExecContext(ctx, "INSERT INTO book_read_link (book_id, user_id, read_status, last_modified, last_time_started_reading, times_started_reading) VALUES (?, ?, ?, ?, ?, 1)", bID, userID, status, now, now)
		}
	}
	return nil
}

// BulkSetArchived atualiza em lote o estado de arquivamento dos livros para o utilizador.
func (d *DB) BulkSetArchived(ctx context.Context, userID int64, bookIDs []int64, archive bool) error {
	if d.appPool == nil || len(bookIDs) == 0 {
		return nil
	}
	now := time.Now().UTC()
	isArchived := 0
	if archive {
		isArchived = 1
	}
	for _, bID := range bookIDs {
		var exists int
		err := d.appPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM archived_book WHERE user_id = ? AND book_id = ?", userID, bID).Scan(&exists)
		if err == nil && exists > 0 {
			_, _ = d.appPool.ExecContext(ctx, "UPDATE archived_book SET is_archived = ?, last_modified = ? WHERE user_id = ? AND book_id = ?", isArchived, now, userID, bID)
		} else {
			_, _ = d.appPool.ExecContext(ctx, "INSERT INTO archived_book (user_id, book_id, is_archived, last_modified) VALUES (?, ?, ?, ?)", userID, bID, isArchived, now)
		}
	}
	return nil
}

// MergeBookRecords mescla múltiplos livros num único registo alvo, transferindo ficheiros e formatos.
func (d *DB) MergeBookRecords(ctx context.Context, targetID int64, sourceIDs []int64, libraryDir string) error {
	target, err := d.GetBookByID(ctx, targetID)
	if err != nil || target == nil {
		return fmt.Errorf("livro de destino não encontrado: %d", targetID)
	}

	targetFormats := make(map[string]bool)
	for _, f := range target.Formats {
		targetFormats[strings.ToUpper(f.Format)] = true
	}

	targetDir := filepath.Join(libraryDir, filepath.FromSlash(target.Path))

	for _, sID := range sourceIDs {
		if sID == targetID {
			continue
		}
		source, err := d.GetBookByID(ctx, sID)
		if err != nil || source == nil {
			continue
		}

		sourceDir := filepath.Join(libraryDir, filepath.FromSlash(source.Path))

		for _, sf := range source.Formats {
			fmtUpper := strings.ToUpper(sf.Format)
			if !targetFormats[fmtUpper] {
				srcFileName := fmt.Sprintf("%s.%s", sf.Name, strings.ToLower(sf.Format))
				dstFileName := fmt.Sprintf("%s.%s", target.Title, strings.ToLower(sf.Format))
				srcFilePath := filepath.Join(sourceDir, srcFileName)
				dstFilePath := filepath.Join(targetDir, dstFileName)

				// Move ou copia o ficheiro físico
				if err := os.Rename(srcFilePath, dstFilePath); err != nil {
					if in, errO := os.Open(srcFilePath); errO == nil {
						if out, errC := os.Create(dstFilePath); errC == nil {
							_, _ = io.Copy(out, in)
							out.Close()
						}
						in.Close()
						_ = os.Remove(srcFilePath)
					}
				}

				// Regista novo formato no livro alvo
				cleanName := strings.TrimSuffix(dstFileName, "."+strings.ToLower(sf.Format))
				_ = d.ExecuteWrite(func(tx *sql.Tx) error {
					_, err := tx.ExecContext(ctx, `
						INSERT INTO data (book, format, uncompressed_size, name)
						VALUES (?, ?, ?, ?)
					`, targetID, fmtUpper, sf.Size, cleanName)
					return err
				})
				targetFormats[fmtUpper] = true
			}
		}

		// Remove o livro de origem do metadata.db e limpa a pasta física
		_ = d.DeleteBook(ctx, sID)
		_ = os.RemoveAll(sourceDir)
	}
	return nil
}

// ReorderShelfBooks atualiza a ordem dos livros numa estante.
func (d *DB) ReorderShelfBooks(ctx context.Context, shelfID int64, orderMap map[int64]int) error {
	if d.appPool == nil {
		return nil
	}
	for bookID, ord := range orderMap {
		_, _ = d.appPool.ExecContext(ctx, "UPDATE book_shelf_link SET \"order\" = ? WHERE shelf = ? AND book_id = ?", ord, shelfID, bookID)
	}
	_, _ = d.appPool.ExecContext(ctx, "UPDATE shelf SET last_modified = CURRENT_TIMESTAMP WHERE id = ?", shelfID)
	return nil
}

func parseFormats(str string) []DataFormat {
	if str == "" {
		return nil
	}
	items := strings.Split(str, ",")
	res := make([]DataFormat, 0, len(items))
	for _, it := range items {
		var parts []string
		if strings.Contains(it, "::") {
			parts = strings.Split(it, "::")
		} else {
			parts = strings.Split(it, ":")
		}
		if len(parts) >= 3 {
			size, _ := strconv.ParseInt(parts[1], 10, 64)
			name := strings.ReplaceAll(parts[2], "__COMMA__", ",")
			res = append(res, DataFormat{
				Format: parts[0],
				Size:   size,
				Name:   name,
			})
		}
	}
	return res
}

func parseCalibreDate(s string) time.Time {
	formats := []string{
		"2006-01-02 15:04:05.999999-07:00",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Now().UTC()
}

// GetSMTPSettings lê a configuração de envio de email do app.db.
func (d *DB) GetSMTPSettings(ctx context.Context) (*SMTPSettings, error) {
	if d.appPool == nil {
		return nil, errors.New("app.db não inicializado")
	}

	var s SMTPSettings
	err := d.appPool.QueryRowContext(ctx, `
		SELECT COALESCE(mail_server, ''), COALESCE(mail_port, 0), COALESCE(mail_use_ssl, 0),
		       COALESCE(mail_login, ''), COALESCE(mail_password, ''), COALESCE(mail_from, '')
		FROM settings
		LIMIT 1
	`).Scan(&s.Server, &s.Port, &s.UseSSL, &s.Login, &s.Password, &s.From)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// EnsureKOSyncTable garante a existência da tabela de sincronização do KOReader no app.db.
func (d *DB) EnsureKOSyncTable(ctx context.Context) error {
	if d.appPool == nil {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS kosync_progress (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			document VARCHAR(255) NOT NULL,
			progress VARCHAR(255) NOT NULL,
			percentage REAL DEFAULT 0,
			device VARCHAR(255),
			device_id VARCHAR(255),
			timestamp INTEGER NOT NULL,
			UNIQUE(user_id, document)
		);
		CREATE INDEX IF NOT EXISTS idx_kosync_user_doc ON kosync_progress(user_id, document);
	`)
	return err
}

// GetKOSyncProgress recupera o registo de leitura do KOReader para o documento e utilizador.
func (d *DB) GetKOSyncProgress(ctx context.Context, userID int64, document string) (*KOSyncProgress, error) {
	if d.appPool == nil {
		return nil, sql.ErrNoRows
	}
	_ = d.EnsureKOSyncTable(ctx)

	var p KOSyncProgress
	err := d.appPool.QueryRowContext(ctx, `
		SELECT user_id, document, progress, percentage, COALESCE(device, ''), COALESCE(device_id, ''), timestamp
		FROM kosync_progress
		WHERE user_id = ? AND document = ?
	`, userID, document).Scan(&p.UserID, &p.Document, &p.Progress, &p.Percentage, &p.Device, &p.DeviceID, &p.Timestamp)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UpdateKOSyncProgress atualiza o progresso de leitura sincronizado do KOReader.
func (d *DB) UpdateKOSyncProgress(ctx context.Context, p KOSyncProgress) error {
	if d.appPool == nil {
		return nil
	}
	_ = d.EnsureKOSyncTable(ctx)

	_, err := d.appPool.ExecContext(ctx, `
		INSERT INTO kosync_progress (user_id, document, progress, percentage, device, device_id, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, document) DO UPDATE SET
			progress = excluded.progress,
			percentage = excluded.percentage,
			device = excluded.device,
			device_id = excluded.device_id,
			timestamp = excluded.timestamp
	`, p.UserID, p.Document, p.Progress, p.Percentage, p.Device, p.DeviceID, p.Timestamp)
	return err
}

// EnsureRemoteAuthTokenTable garante a existência da tabela remote_auth_token no app.db.
func (d *DB) EnsureRemoteAuthTokenTable(ctx context.Context) error {
	if d.appPool == nil {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS remote_auth_token (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			auth_token VARCHAR(255) UNIQUE,
			user_id INTEGER NOT NULL,
			verified BOOLEAN DEFAULT 0,
			expiration DATETIME,
			token_type INTEGER DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_remote_auth_token ON remote_auth_token(auth_token, token_type);
	`)
	return err
}

// GetUserByKoboToken pesquisa o utilizador associado a um token do Kobo.
func (d *DB) GetUserByKoboToken(ctx context.Context, token string) (*User, error) {
	if d.appPool == nil {
		return nil, errors.New("app.db não inicializado")
	}
	_ = d.EnsureRemoteAuthTokenTable(ctx)

	var u User
	err := d.appPool.QueryRowContext(ctx, `
		SELECT u.id, u.name, COALESCE(u.email, ''), COALESCE(u.kindle_mail, ''),
		       u.role, COALESCE(u.locale, 'en'), COALESCE(u.default_language, 'all'), u.password
		FROM remote_auth_token r
		JOIN user u ON r.user_id = u.id
		WHERE r.auth_token = ? AND r.token_type = 1
		LIMIT 1
	`, token).Scan(&u.ID, &u.Name, &u.Email, &u.KindleMail, &u.Role, &u.Locale, &u.DefaultLanguage, &u.PasswordHash)
	if err != nil {
		// Fallback: se o token for uma session_key válida
		if sessionUser, sErr := d.GetUserBySessionKey(ctx, token); sErr == nil {
			return sessionUser, nil
		}
		return nil, err
	}
	u.IsAdmin = (u.Role & 1) != 0
	return &u, nil
}

// GetOrCreateKoboToken obtém ou cria um token de sincronização Kobo para o utilizador.
func (d *DB) GetOrCreateKoboToken(ctx context.Context, userID int64) (string, error) {
	if d.appPool == nil {
		return "", errors.New("app.db não inicializado")
	}
	_ = d.EnsureRemoteAuthTokenTable(ctx)

	var token string
	err := d.appPool.QueryRowContext(ctx, `
		SELECT auth_token FROM remote_auth_token WHERE user_id = ? AND token_type = 1 LIMIT 1
	`, userID).Scan(&token)
	if err == nil && token != "" {
		return token, nil
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token = hex.EncodeToString(b)

	_, err = d.appPool.ExecContext(ctx, `
		INSERT INTO remote_auth_token (auth_token, user_id, verified, expiration, token_type)
		VALUES (?, ?, 1, '9999-12-31 23:59:59', 1)
	`, token, userID)
	if err != nil {
		return "", err
	}
	return token, nil
}
