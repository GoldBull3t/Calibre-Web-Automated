package formats_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crocodilestick/calibre-web-go/internal/formats"
)

func TestExtractEPUB(t *testing.T) {
	epubPath := filepath.Join("..", "..", "tests", "fixtures", "sample_books", "alice_in_wonderland.epub")
	book, err := formats.Extract(epubPath)
	if err != nil {
		t.Fatalf("Erro ao extrair EPUB: %v", err)
	}

	if book.Title == "" {
		t.Fatal("Título não pode ser vazio")
	}
	if len(book.Authors) == 0 {
		t.Fatal("Autores não podem ser vazios")
	}
	if book.Format != "EPUB" {
		t.Fatalf("Formato esperado EPUB, obtido: %s", book.Format)
	}
}

func TestExtractMOBI(t *testing.T) {
	mobiPath := filepath.Join("..", "..", "tests", "fixtures", "sample_books", "alice_in_wonderland.mobi")
	book, err := formats.Extract(mobiPath)
	if err != nil {
		t.Fatalf("Erro ao extrair MOBI: %v", err)
	}

	if book.Title == "" {
		t.Fatal("Título não pode ser vazio")
	}
	if book.Format != "MOBI" {
		t.Fatalf("Formato esperado MOBI, obtido: %s", book.Format)
	}
}

func TestExtractTXT(t *testing.T) {
	txtPath := filepath.Join("..", "..", "tests", "fixtures", "sample_books", "alice_in_wonderland.txt")
	book, err := formats.Extract(txtPath)
	if err != nil {
		t.Fatalf("Erro ao extrair TXT: %v", err)
	}

	if book.Title == "" {
		t.Fatal("Título não pode ser vazio")
	}
	if book.Format != "TXT" {
		t.Fatalf("Formato esperado TXT, obtido: %s", book.Format)
	}
}

func TestFixEPUB(t *testing.T) {
	orig := filepath.Join("..", "..", "tests", "fixtures", "sample_books", "alice_in_wonderland.epub")
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.epub")

	data, err := os.ReadFile(orig)
	if err != nil {
		t.Fatalf("Erro ao ler EPUB de teste: %v", err)
	}
	if err := os.WriteFile(target, data, 0644); err != nil {
		t.Fatalf("Erro ao gravar cópia temporária de EPUB: %v", err)
	}

	fixes, err := formats.FixEPUB(target)
	if err != nil {
		t.Fatalf("Erro ao executar FixEPUB: %v", err)
	}
	t.Logf("FixEPUB executado com sucesso: %d correções aplicadas", fixes)

	// Valida se o EPUB corrigido continua legível e íntegro
	book, err := formats.Extract(target)
	if err != nil {
		t.Fatalf("EPUB corrigido corrompido: %v", err)
	}
	if book.Title == "" {
		t.Fatal("Título do EPUB corrigido não pode estar vazio")
	}
}

func TestCleanOCRErrors(t *testing.T) {
	inputHTML := []byte(`<p class="sample">Este é um n0vo l1vro com ﬁm fascinante .</p><p>Quebra de pa-
lavra e hífens com espaçamentos , que devem ser corrigidos.</p><img src="cover01.jpg" alt="img0" />`)

	cleaned, fixes := formats.CleanOCRErrors(inputHTML)
	if fixes == 0 {
		t.Fatalf("Esperado que fossem encontradas correções de OCR, obteve 0")
	}

	cleanedStr := string(cleaned)

	// Valida tags HTML intocadas
	if !strings.Contains(cleanedStr, `src="cover01.jpg"`) {
		t.Errorf("Tag HTML foi corrompida: %s", cleanedStr)
	}

	// Valida substituições de texto
	if !strings.Contains(cleanedStr, "novo") {
		t.Errorf("n0vo não foi corrigido para novo: %s", cleanedStr)
	}
	if !strings.Contains(cleanedStr, "livro") {
		t.Errorf("l1vro não foi corrigido para livro: %s", cleanedStr)
	}
	if !strings.Contains(cleanedStr, "fim") {
		t.Errorf("ligadura ﬁm não foi corrigida para fim: %s", cleanedStr)
	}
	if !strings.Contains(cleanedStr, "palavra") {
		t.Errorf("pa-\\nlavra não foi unificada: %s", cleanedStr)
	}
	if strings.Contains(cleanedStr, " , ") {
		t.Errorf("Espaço antes da vírgula não foi corrigido: %s", cleanedStr)
	}
	if strings.Contains(cleanedStr, " .") {
		t.Errorf("Espaço antes do ponto final não foi corrigido: %s", cleanedStr)
	}
}

func TestComicPages(t *testing.T) {
	tmpDir := t.TempDir()
	cbzPath := filepath.Join(tmpDir, "test_comic.cbz")

	zipFile, err := os.Create(cbzPath)
	if err != nil {
		t.Fatalf("Erro ao criar cbz temporário: %v", err)
	}
	zw := zip.NewWriter(zipFile)
	w1, _ := zw.Create("01_cover.jpg")
	_, _ = w1.Write([]byte("mock_img_1"))
	w2, _ := zw.Create("02_page.jpg")
	_, _ = w2.Write([]byte("mock_img_2"))
	_ = zw.Close()
	_ = zipFile.Close()

	info, err := formats.ListComicPages(cbzPath)
	if err != nil {
		t.Fatalf("Erro ao listar páginas do comic: %v", err)
	}
	if info.TotalPages != 2 {
		t.Fatalf("Esperadas 2 páginas, obtidas %d", info.TotalPages)
	}
	if info.Pages[0] != "01_cover.jpg" || info.Pages[1] != "02_page.jpg" {
		t.Fatalf("Ordem natural de páginas incorreta: %v", info.Pages)
	}

	data, mimeType, err := formats.ExtractComicPage(cbzPath, 0)
	if err != nil {
		t.Fatalf("Erro ao extrair página 0: %v", err)
	}
	if len(data) != 10 {
		t.Fatalf("Tamanho inesperado de bytes de imagem: %d", len(data))
	}
	if mimeType != "image/jpeg" {
		t.Fatalf("MIME type esperado image/jpeg, obtido: %s", mimeType)
	}
}

