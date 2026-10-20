package search

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestFTS5Search(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir sqlite in-memory: %v", err)
	}
	defer db.Close()

	idx, err := NewIndexer(db, ".")
	if err != nil {
		t.Fatalf("falha ao criar indexer FTS5: %v", err)
	}

	ctx := context.Background()

	// Insere documento de teste manualmente
	_, err = db.ExecContext(ctx, `
		INSERT INTO books_fts (book_id, title, author, content)
		VALUES (?, ?, ?, ?)
	`, 42, "Sherlock Holmes", "Arthur Conan Doyle", "Elementary, my dear Watson. The game is afoot in London Baker Street.")
	if err != nil {
		t.Fatalf("falha ao inserir no FTS5: %v", err)
	}

	// Pesquisa por 'Watson'
	results, err := idx.Search(ctx, "Watson", 10)
	if err != nil {
		t.Fatalf("erro na pesquisa FTS5: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("esperava 1 resultado, obteve %d", len(results))
	}

	r := results[0]
	if r.BookID != 42 {
		t.Errorf("esperava book_id 42, obteve %d", r.BookID)
	}

	if !stringsContains(r.Snippet, "<mark>Watson</mark>") {
		t.Errorf("esperava snippet com marcação Watson, obteve: %s", r.Snippet)
	}
}

func stringsContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSub(s, sub))
}

func containsSub(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

