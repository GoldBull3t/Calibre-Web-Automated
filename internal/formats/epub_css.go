package formats

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	// Remove regras rígidas de font-family
	cssFontFamilyRegex = regexp.MustCompile(`(?i)font-family\s*:\s*[^;!]+(?:!important)?\s*;?`)

	// Remove cores de texto pretas ou escuras forçadas que estragam o Dark Mode
	cssForcedColorRegex = regexp.MustCompile(`(?i)(?:^|[\s;{])color\s*:\s*(?:#000000|#000|#111111|#111|black|rgb\(\s*0\s*,\s*0\s*,\s*0\s*\))\s*(?:!important)?\s*;?`)

	// Remove cores de fundo brancas forçadas que estragam o Dark Mode
	cssForcedBgRegex = regexp.MustCompile(`(?i)(?:^|[\s;{])background(?:-color)?\s*:\s*(?:#ffffff|#fff|#fafafa|white|rgb\(\s*255\s*,\s*255\s*,\s*255\s*\))\s*(?:!important)?\s*;?`)

	// Remove margens e paddings fixos excessivos em body ou @page
	cssFixedMarginsRegex = regexp.MustCompile(`(?i)(?:margin|padding)(?:-left|-right)?\s*:\s*(?:[3-9]\d|\d{3,})px\s*(?:!important)?\s*;?`)
)

// SanitizeEPUBContent limpa CSS e estilos inline agressivos de um ficheiro de texto (HTML, XHTML ou CSS).
func SanitizeEPUBContent(data []byte, filename string) ([]byte, int) {
	lowerName := strings.ToLower(filename)
	isCSS := strings.HasSuffix(lowerName, ".css")
	isHTML := strings.HasSuffix(lowerName, ".xhtml") || strings.HasSuffix(lowerName, ".html") || strings.HasSuffix(lowerName, ".htm")

	if !isCSS && !isHTML {
		return data, 0
	}

	str := string(data)
	origLen := len(str)

	// Limpa cores forçadas
	str = cssForcedColorRegex.ReplaceAllString(str, " ")
	str = cssForcedBgRegex.ReplaceAllString(str, " ")

	// Limpa fontes rígidas
	str = cssFontFamilyRegex.ReplaceAllString(str, " ")

	// Limpa margens excessivas
	str = cssFixedMarginsRegex.ReplaceAllString(str, " ")

	fixes := 0
	if len(str) != origLen {
		fixes = 1
	}

	return []byte(str), fixes
}

// SanitizeEPUB limpa o CSS/HTML de um ficheiro EPUB preservando a especificação IDPF
// (o ficheiro mimetype DEVE ser gravado sem compressão como primeira entrada do ZIP).
func SanitizeEPUB(epubPath, outputPath string) (int, error) {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return 0, fmt.Errorf("falha ao abrir EPUB para sanitização: %w", err)
	}
	defer r.Close()

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	totalFixes := 0

	// 1. O primeiro ficheiro DEVE ser o 'mimetype' com método Store (sem compressão), conforme especificação IDPF
	mimetypeWritten := false
	for _, f := range r.File {
		if f.Name == "mimetype" {
			rc, err := f.Open()
			if err != nil {
				w.Close()
				return 0, err
			}
			mimeBytes, _ := io.ReadAll(rc)
			rc.Close()

			hdr := &zip.FileHeader{
				Name:   "mimetype",
				Method: zip.Store, // Sem compressão obrigatório
			}
			writer, err := w.CreateHeader(hdr)
			if err != nil {
				w.Close()
				return 0, err
			}
			_, _ = writer.Write(mimeBytes)
			mimetypeWritten = true
			break
		}
	}

	// Se não existia mimetype explícito, cria-o conforme o standard
	if !mimetypeWritten {
		hdr := &zip.FileHeader{
			Name:   "mimetype",
			Method: zip.Store,
		}
		writer, err := w.CreateHeader(hdr)
		if err == nil {
			_, _ = writer.Write([]byte("application/epub+zip"))
		}
	}

	// 2. Copia e sanitiza os restantes ficheiros com Deflate
	for _, f := range r.File {
		if f.Name == "mimetype" {
			continue // Já processado
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}

		sanitizedData, fixes := SanitizeEPUBContent(data, f.Name)
		totalFixes += fixes

		hdr := &zip.FileHeader{
			Name:   f.Name,
			Method: zip.Deflate,
		}
		writer, err := w.CreateHeader(hdr)
		if err != nil {
			continue
		}
		_, _ = writer.Write(sanitizedData)
	}

	if err := w.Close(); err != nil {
		return 0, fmt.Errorf("falha ao fechar escritor ZIP: %w", err)
	}

	// Grava no ficheiro de saída
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return 0, fmt.Errorf("falha ao gravar EPUB sanitizado: %w", err)
	}

	return totalFixes, nil
}
