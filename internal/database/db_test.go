package database

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func TestQueries(t *testing.T) {
	// Create a temporary sqlite metadata.db with mock data
	tmpDir, err := os.MkdirTemp("", "calibre-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	metaDb := tmpDir + "/metadata.db"
	appDb := tmpDir + "/app.db"

	db, err := Open(metaDb, appDb)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create tables matching calibre
	err = db.ExecuteWrite(func(tx *sql.Tx) error {
		_, err := tx.Exec(`
			CREATE TABLE authors (id INTEGER PRIMARY KEY, name TEXT NOT NULL, sort TEXT, link TEXT);
			CREATE TABLE books (id INTEGER PRIMARY KEY, title TEXT NOT NULL, sort TEXT, timestamp TEXT, pubdate TEXT, series_index REAL, last_modified TEXT, path TEXT, has_cover INTEGER, uuid TEXT);
			CREATE TABLE books_authors_link (id INTEGER PRIMARY KEY, book INTEGER NOT NULL, author INTEGER NOT NULL);
			CREATE TABLE tags (id INTEGER PRIMARY KEY, name TEXT NOT NULL);
			CREATE TABLE books_tags_link (id INTEGER PRIMARY KEY, book INTEGER NOT NULL, tag INTEGER NOT NULL);
			CREATE TABLE series (id INTEGER PRIMARY KEY, name TEXT NOT NULL, sort TEXT);
			CREATE TABLE books_series_link (id INTEGER PRIMARY KEY, book INTEGER NOT NULL, series INTEGER NOT NULL);
			CREATE TABLE data (id INTEGER PRIMARY KEY, book INTEGER NOT NULL, format TEXT NOT NULL, uncompressed_size INTEGER, name TEXT);
		`)
		if err != nil {
			return err
		}

		// Insert 2 authors, 2 books, 2 tags, 2 formats
		_, _ = tx.Exec("INSERT INTO authors (id, name) VALUES (1, 'Fernando Pessoa'), (2, 'José Saramago')")
		_, _ = tx.Exec("INSERT INTO books (id, title, sort) VALUES (1, 'Mensagem', 'Mensagem'), (2, 'Ensaio', 'Ensaio')")
		_, _ = tx.Exec("INSERT INTO books_authors_link (book, author) VALUES (1, 1), (2, 2)")
		_, _ = tx.Exec("INSERT INTO tags (id, name) VALUES (1, 'Poesia'), (2, 'Ficção')")
		_, _ = tx.Exec("INSERT INTO books_tags_link (book, tag) VALUES (1, 1), (2, 2)")
		_, _ = tx.Exec("INSERT INTO data (book, format, uncompressed_size, name) VALUES (1, 'EPUB', 1000, 'Mensagem - FP'), (2, 'PDF', 2000, 'Ensaio - JS')")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	authors, err := db.GetAllAuthors(ctx)
	if err != nil {
		t.Fatalf("GetAllAuthors failed: %v", err)
	}
	t.Logf("Authors: %+v", authors)

	tags, err := db.GetAllTags(ctx)
	if err != nil {
		t.Fatalf("GetAllTags failed: %v", err)
	}
	t.Logf("Tags: %+v", tags)

	formats, err := db.ListFormats(ctx)
	if err != nil {
		t.Fatalf("ListFormats failed: %v", err)
	}
	t.Logf("Formats: %+v", formats)
}
