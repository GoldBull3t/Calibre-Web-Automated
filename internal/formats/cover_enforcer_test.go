package formats_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/crocodilestick/calibre-web-go/internal/formats"
)

func TestEnforceCoverAndMetadata(t *testing.T) {
	orig := filepath.Join("..", "..", "tests", "fixtures", "sample_books", "alice_in_wonderland.epub")
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "enforce_test.epub")

	data, err := os.ReadFile(orig)
	if err != nil {
		t.Fatalf("Erro ao ler EPUB fixture: %v", err)
	}
	if err := os.WriteFile(target, data, 0644); err != nil {
		t.Fatalf("Erro ao copiar EPUB: %v", err)
	}

	fakeCover := filepath.Join(tmpDir, "test_cover.jpg")
	if err := os.WriteFile(fakeCover, []byte("JPEG_FAKE_IMAGE_DATA_12345"), 0644); err != nil {
		t.Fatalf("Erro ao gravar capa de teste: %v", err)
	}

	newTitle := "Alice no Pais das Maravilhas Enforced"
	newAuthor := "Lewis Carroll CWA"

	err = formats.EnforceCoverAndMetadata(target, fakeCover, newTitle, newAuthor)
	if err != nil {
		t.Fatalf("EnforceCoverAndMetadata falhou: %v", err)
	}

	// Extrai novamente os metadados do EPUB para garantir que a alteração física ocorreu
	updatedBook, err := formats.Extract(target)
	if err != nil {
		t.Fatalf("Erro ao extrair EPUB após enforcement: %v", err)
	}

	if updatedBook.Title != newTitle {
		t.Fatalf("Esperado título %q, obtido %q", newTitle, updatedBook.Title)
	}
	if len(updatedBook.Authors) == 0 || updatedBook.Authors[0] != newAuthor {
		t.Fatalf("Esperado autor %q, obtido %v", newAuthor, updatedBook.Authors)
	}
}

