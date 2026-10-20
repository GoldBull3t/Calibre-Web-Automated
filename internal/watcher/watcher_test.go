package watcher

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/crocodilestick/calibre-web-go/internal/formats"
)

func setupWatcherTest(t *testing.T) (*database.DB, *Watcher, string, func()) {
	tmpDir, err := os.MkdirTemp("", "cwa-watcher-test-*")
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

	ingestDir := filepath.Join(tmpDir, "ingest")
	libraryDir := filepath.Join(tmpDir, "library")
	_ = os.MkdirAll(ingestDir, 0755)
	_ = os.MkdirAll(libraryDir, 0755)

	w := New(ingestDir, libraryDir, db)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, w, tmpDir, cleanup
}

func TestWatcherMultiFormatAndDeduplication(t *testing.T) {
	db, w, tmpDir, cleanup := setupWatcherTest(t)
	defer cleanup()

	sampleDir := filepath.Join("..", "..", "tests", "fixtures", "sample_books")
	epubSrc := filepath.Join(sampleDir, "alice_in_wonderland.epub")
	mobiSrc := filepath.Join(sampleDir, "alice_in_wonderland.mobi")
	txtSrc := filepath.Join(sampleDir, "alice_in_wonderland.txt")

	// 1. Ingesta EPUB inicial
	epubTest := filepath.Join(tmpDir, "ingest", "alice.epub")
	data, _ := os.ReadFile(epubSrc)
	_ = os.WriteFile(epubTest, data, 0644)

	initialTotal, _, _, _ := db.GetStats(context.Background())

	if err := w.processFile(epubTest); err != nil {
		t.Fatalf("Erro ao processar EPUB: %v", err)
	}

	// Verifica se novo livro foi adicionado
	totalAfterEPUB, _, _, _ := db.GetStats(context.Background())
	if totalAfterEPUB != initialTotal+1 {
		t.Fatalf("Esperado %d livros após EPUB, obtido %d", initialTotal+1, totalAfterEPUB)
	}

	res, err := db.ListBooks(context.Background(), database.BookFilter{Page: 1, PageSize: 10, Sort: "new"})
	if err != nil || len(res.Items) == 0 {
		t.Fatalf("Nenhum livro retornado: %v", err)
	}
	latestBook := res.Items[0]

	// 2. Ingesta MOBI com mesmo título e autor -> DEVE ANEXAR FORMATO sem criar novo livro!
	mobiTest := filepath.Join(tmpDir, "ingest", "alice.mobi")
	mobiData, _ := os.ReadFile(mobiSrc)
	_ = os.WriteFile(mobiTest, mobiData, 0644)

	if err := w.processFile(mobiTest); err != nil {
		t.Fatalf("Erro ao processar MOBI: %v", err)
	}

	// 3. Ingesta TXT -> DEVE ANEXAR FORMATO sem criar novo livro!
	txtTest := filepath.Join(tmpDir, "ingest", "Alice's Adventures in Wonderland - Lewis Carroll.txt")
	txtData, _ := os.ReadFile(txtSrc)
	_ = os.WriteFile(txtTest, txtData, 0644)

	if err := w.processFile(txtTest); err != nil {
		t.Fatalf("Erro ao processar TXT: %v", err)
	}

	// Verifica total de livros na biblioteca: NÃO DEVE TER AUMENTADO!
	finalTotal, _, _, _ := db.GetStats(context.Background())
	if finalTotal != initialTotal+1 {
		t.Fatalf("Deduplicação falhou: esperava %d livros total com múltiplos formatos, obtido %d", initialTotal+1, finalTotal)
	}

	updatedBook, _ := db.GetBookByID(context.Background(), latestBook.ID)
	if updatedBook == nil || len(updatedBook.Formats) < 2 {
		t.Fatalf("Esperado múltiplos formatos anexados ao livro %d, obtido: %v", latestBook.ID, updatedBook)
	}
}

func TestAutoZipperUnpack(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cwa-zip-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "my_books.zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Failed to create zip: %v", err)
	}

	zw := zip.NewWriter(zf)
	f1, _ := zw.Create("book1.epub")
	_, _ = f1.Write([]byte("fake epub content"))
	f2, _ := zw.Create("ignore_me.nfo")
	_, _ = f2.Write([]byte("fake nfo content"))
	_ = zw.Close()
	_ = zf.Close()

	destDir := filepath.Join(tmpDir, "unpacked")
	_ = os.MkdirAll(destDir, 0755)

	extracted, err := formats.UnpackZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("Erro no UnpackZip: %v", err)
	}

	if len(extracted) != 1 {
		t.Fatalf("Esperado 1 ebook extraído do zip (.epub), obtido: %d", len(extracted))
	}
}
