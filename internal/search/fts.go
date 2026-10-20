package search

import (
	"archive/zip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var tagRegex = regexp.MustCompile(`<[^>]*>`)
var wsRegex = regexp.MustCompile(`\s+`)

type FTSResult struct {
	BookID  int64  `json:"book_id"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Snippet string `json:"snippet"`
}

type Indexer struct {
	db       *sql.DB
	booksDir string
	mu       sync.Mutex
}

func NewIndexer(db *sql.DB, booksDir string) (*Indexer, error) {
	idx := &Indexer{
		db:       db,
		booksDir: booksDir,
	}

	if err := idx.ensureTable(); err != nil {
		return nil, fmt.Errorf("falha ao inicializar FTS5: %w", err)
	}

	return idx, nil
}

func (idx *Indexer) ensureTable() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	_, err := idx.db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS books_fts USING fts5(
			book_id UNINDEXED,
			title,
			author,
			content,
			tokenize = 'porter unicode61'
		);
	`)
	return err
}

func (idx *Indexer) IndexBookEPUB(ctx context.Context, bookID int64, title, author, relPath, fileName string) error {
	fullPath := filepath.Join(idx.booksDir, filepath.FromSlash(relPath), fileName)
	text, err := extractEPUBText(fullPath)
	if err != nil {
		return fmt.Errorf("erro ao extrair texto do epub %s: %w", fullPath, err)
	}

	if len(strings.TrimSpace(text)) == 0 {
		return nil
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	_, _ = idx.db.ExecContext(ctx, "DELETE FROM books_fts WHERE book_id = ?", bookID)

	_, err = idx.db.ExecContext(ctx, `
		INSERT INTO books_fts (book_id, title, author, content)
		VALUES (?, ?, ?, ?)
	`, bookID, title, author, text)

	return err
}

func (idx *Indexer) Search(ctx context.Context, query string, limit int) ([]FTSResult, error) {
	if limit <= 0 {
		limit = 30
	}

	cleanQuery := sanitizeFTSQuery(query)
	if cleanQuery == "" {
		return nil, nil
	}

	rows, err := idx.db.QueryContext(ctx, `
		SELECT book_id, title, author, snippet(books_fts, 3, '<mark>', '</mark>', '...', 25)
		FROM books_fts
		WHERE books_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, cleanQuery, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []FTSResult
	for rows.Next() {
		var r FTSResult
		if err := rows.Scan(&r.BookID, &r.Title, &r.Author, &r.Snippet); err != nil {
			continue
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func extractEPUBText(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var sb strings.Builder
	for _, f := range r.File {
		name := strings.ToLower(f.Name)
		if strings.HasSuffix(name, ".xhtml") || strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".htm") || strings.HasSuffix(name, ".txt") {
			rc, err := f.Open()
			if err != nil {
				continue
			}

			// Limite suave de leitura por capítulo (500KB) para manter memória sob controlo
			lr := io.LimitReader(rc, 500*1024)
			b, err := io.ReadAll(lr)
			_ = rc.Close()
			if err == nil {
				clean := tagRegex.ReplaceAllString(string(b), " ")
				clean = wsRegex.ReplaceAllString(clean, " ")
				sb.WriteString(clean)
				sb.WriteString("\n")
			}
		}
	}

	return sb.String(), nil
}

func sanitizeFTSQuery(q string) string {
	q = strings.TrimSpace(q)
	// Remove caracteres com significado especial no FTS5 se não balanceados
	replacer := strings.NewReplacer(`"`, "", `'`, "", `*`, "", `:`, "", `(`, "", `)`, "", `{`, "", `}`, "", `^`, "")
	q = replacer.Replace(q)
	words := strings.Fields(q)
	if len(words) == 0 {
		return ""
	}
	// Junta palavras como prefix match
	var parts []string
	for _, w := range words {
		parts = append(parts, fmt.Sprintf(`"%s"*`, w))
	}
	return strings.Join(parts, " AND ")
}

