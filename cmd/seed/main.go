package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/crocodilestick/calibre-web-go/internal/epub"
	"github.com/crocodilestick/calibre-web-go/internal/watcher"
)

type SeedBook struct {
	GutenbergID int
	Filename    string
	Title       string
	Author      string
}

var realBooks = []SeedBook{
	{GutenbergID: 84, Filename: "frankenstein.epub", Title: "Frankenstein; Or, The Modern Prometheus", Author: "Mary Wollstonecraft Shelley"},
	{GutenbergID: 345, Filename: "dracula.epub", Title: "Dracula", Author: "Bram Stoker"},
	{GutenbergID: 174, Filename: "dorian_gray.epub", Title: "The Picture of Dorian Gray", Author: "Oscar Wilde"},
	{GutenbergID: 64317, Filename: "great_gatsby.epub", Title: "The Great Gatsby", Author: "F. Scott Fitzgerald"},
	{GutenbergID: 43, Filename: "jekyll_and_hyde.epub", Title: "The Strange Case of Dr. Jekyll and Mr. Hyde", Author: "Robert Louis Stevenson"},
	{GutenbergID: 35, Filename: "time_machine.epub", Title: "The Time Machine", Author: "H. G. Wells"},
	{GutenbergID: 5200, Filename: "metamorphosis.epub", Title: "The Metamorphosis", Author: "Franz Kafka"},
	{GutenbergID: 1952, Filename: "yellow_wallpaper.epub", Title: "The Yellow Wallpaper", Author: "Charlotte Perkins Gilman"},
	{GutenbergID: 1232, Filename: "the_prince.epub", Title: "The Prince", Author: "Niccolò Machiavelli"},
	{GutenbergID: 132, Filename: "art_of_war.epub", Title: "The Art of War", Author: "Sun Tzu"},
}

func main() {
	cacheDir := filepath.Join("tests", "fixtures", "sample_books")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("Erro ao criar cacheDir: %v", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}

	// 1. Descarrega os 10 livros reais de teste
	log.Println("[SEED] A preparar 10 livros reais clássicos com capas e texto integral...")
	for _, b := range realBooks {
		destPath := filepath.Join(cacheDir, b.Filename)
		if fi, err := os.Stat(destPath); err == nil && fi.Size() > 10000 {
			log.Printf("[CACHE] %s já descarregado (%d KB)", b.Title, fi.Size()/1024)
			continue
		}

		url := fmt.Sprintf("https://www.gutenberg.org/cache/epub/%d/pg%d-images.epub", b.GutenbergID, b.GutenbergID)
		log.Printf("[DOWNLOAD] A descarregar '%s' de %s...", b.Title, url)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			log.Fatalf("Erro ao criar requisição: %v", err)
		}
		req.Header.Set("User-Agent", "Calibre-Web-Automated-Seeder/1.0 (https://github.com/crocodilestick)")

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Fatalf("Falha no download de %s (status %v): %v", b.Title, resp.StatusCode, err)
		}

		out, err := os.Create(destPath)
		if err != nil {
			resp.Body.Close()
			log.Fatalf("Erro ao criar ficheiro %s: %v", destPath, err)
		}
		_, err = io.Copy(out, resp.Body)
		resp.Body.Close()
		out.Close()

		if err != nil {
			log.Fatalf("Erro ao gravar %s: %v", destPath, err)
		}
		log.Printf("[SUCCESS] '%s' descarregado com sucesso!", b.Title)
	}

	// 2. Ingestão para os diretórios de biblioteca
	targetLibraries := []string{"empty_library", "library"}
	tmpIngest := filepath.Join("ingest", "seed_temp")
	_ = os.MkdirAll(tmpIngest, 0755)
	defer os.RemoveAll(tmpIngest)

	for _, libDir := range targetLibraries {
		dbPath := filepath.Join(libDir, "metadata.db")
		appDbPath := filepath.Join(libDir, "app.db")

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			continue
		}

		log.Printf("\n[IMPORT] A importar livros para biblioteca: %s", libDir)
		db, err := database.Open(dbPath, appDbPath)
		if err != nil {
			log.Printf("[WARN] Não foi possível abrir %s: %v", dbPath, err)
			continue
		}

		w := watcher.New(tmpIngest, libDir, db)

		for _, b := range realBooks {
			srcPath := filepath.Join(cacheDir, b.Filename)
			tempPath := filepath.Join(tmpIngest, b.Filename)

			// Faz cópia temporária para o watcher processar
			if err := copyFile(srcPath, tempPath); err != nil {
				log.Printf("[ERROR] Erro ao copiar %s para temp: %v", b.Filename, err)
				continue
			}

			if err := w.ProcessFile(tempPath); err != nil {
				log.Printf("[ERROR] Falha ao processar %s: %v", b.Title, err)
			} else {
				log.Printf("  ✓ Livro incorporado: '%s' por %s", b.Title, b.Author)
			}
		}

		// Restaura capa de Charles Dickens caso tenha ficado mock
		fixChristmasCarolCover(libDir, cacheDir, db)

		// Checkpoint WAL para persistência imediata
		db.Close()
		log.Printf("[DONE] Biblioteca %s atualizada e sincronizada com sucesso!", libDir)
	}
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

	_, err = io.Copy(out, in)
	return err
}

func fixChristmasCarolCover(libDir, cacheDir string, db *database.DB) {
	sampleChristmas := filepath.Join(cacheDir, "christmas_carol.epub")
	if _, err := os.Stat(sampleChristmas); err != nil {
		return
	}

	var coverBuf bytes.Buffer
	if err := epub.ExtractCover(sampleChristmas, &coverBuf); err == nil && coverBuf.Len() > 1000 {
		var path string
		err := db.MetadataDB().QueryRow("SELECT path FROM books WHERE title LIKE '%Christmas Carol%'").Scan(&path)
		if err == nil && path != "" {
			coverFile := filepath.Join(libDir, filepath.FromSlash(path), "cover.jpg")
			if fi, err := os.Stat(coverFile); err != nil || fi.Size() < 1000 {
				_ = os.WriteFile(coverFile, coverBuf.Bytes(), 0644)
				_ = db.ExecuteWrite(func(tx *sql.Tx) error {
					_, err := tx.Exec("UPDATE books SET has_cover = 1 WHERE title LIKE '%Christmas Carol%'")
					return err
				})
				log.Printf("  ✓ Capa original de 'A Christmas Carol' restaurada em %s", libDir)
			}
		}
	}
}
