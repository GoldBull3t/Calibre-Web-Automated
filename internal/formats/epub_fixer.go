package formats

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	xmlDeclRegex = regexp.MustCompile(`<\?xml[^>]*\?>`)
	langTagRegex = regexp.MustCompile(`(?i)<dc:language>([^<]*)</dc:language>`)
	strayImgRegex = regexp.MustCompile(`(?i)<img([^>]*[^/])>`)
	strayBrRegex  = regexp.MustCompile(`(?i)<br(\s*)>`)

	ocrHyphenRegex  = regexp.MustCompile(`([\p{L}]{2,})[-–—]\s*[\r\n]+\s*([\p{L}]{2,})`)
	ocrSoftHyphen   = regexp.MustCompile(`\xad|&shy;|&#173;|&#xAD;`)
	ocrDigitZero    = regexp.MustCompile(`\b([\p{L}]+)0([\p{L}]+)\b`)
	ocrDigitOne     = regexp.MustCompile(`\b([\p{L}]+)1([\p{L}]+)\b`)
	ocrPunctSpacing = regexp.MustCompile(`\s+([,;:\!\?\.])`)
	ocrDoubleSpaces = regexp.MustCompile(`[ \t]{2,}`)

	ocrReplacer = strings.NewReplacer(
		"ﬁ", "fi",
		"ﬂ", "fl",
		"ﬀ", "ff",
		"ﬃ", "ffi",
		"ﬄ", "ffl",
		"œ", "oe",
		"æ", "ae",
		"â€œ", "\"",
		"â€\u009d", "\"",
		"â€˜", "'",
		"â€™", "'",
		"â€”", "—",
		"â€“", "–",
	)
)

// CleanOCRErrors limpa artefactos de OCR e digitalização em nós de texto HTML/XHTML.
func CleanOCRErrors(content []byte) ([]byte, int) {
	fixes := 0
	var out bytes.Buffer
	out.Grow(len(content))

	i := 0
	inScript := false
	inStyle := false

	for i < len(content) {
		if content[i] == '<' {
			// Início de tag HTML
			end := bytes.IndexByte(content[i:], '>')
			if end == -1 {
				out.Write(content[i:])
				break
			}
			tag := content[i : i+end+1]
			tagLower := bytes.ToLower(tag)

			if bytes.HasPrefix(tagLower, []byte("<script")) {
				inScript = true
			} else if bytes.HasPrefix(tagLower, []byte("</script")) {
				inScript = false
			} else if bytes.HasPrefix(tagLower, []byte("<style")) {
				inStyle = true
			} else if bytes.HasPrefix(tagLower, []byte("</style")) {
				inStyle = false
			}

			out.Write(tag)
			i += end + 1
		} else {
			// Nó de texto fora de tag
			nextTag := bytes.IndexByte(content[i:], '<')
			var textChunk []byte
			if nextTag == -1 {
				textChunk = content[i:]
				i = len(content)
			} else {
				textChunk = content[i : i+nextTag]
				i += nextTag
			}

			if inScript || inStyle || len(bytes.TrimSpace(textChunk)) == 0 {
				out.Write(textChunk)
				continue
			}

			origLen := len(textChunk)
			s := string(textChunk)

			// 1. Substituição de ligaduras e mojibake
			replaced := ocrReplacer.Replace(s)
			if replaced != s {
				fixes++
				s = replaced
			}

			// 2. Remoção de soft hyphens
			if ocrSoftHyphen.MatchString(s) {
				s = ocrSoftHyphen.ReplaceAllString(s, "")
				fixes++
			}

			// 3. Junção de palavras partidas por quebra de linha com hífen
			if ocrHyphenRegex.MatchString(s) {
				s = ocrHyphenRegex.ReplaceAllString(s, "$1$2")
				fixes++
			}

			// 4. Correção de dígitos dentro de palavras (ex: n0vo -> novo, d1reito -> direito)
			if ocrDigitZero.MatchString(s) {
				s = ocrDigitZero.ReplaceAllString(s, "${1}o${2}")
				fixes++
			}
			if ocrDigitOne.MatchString(s) {
				s = ocrDigitOne.ReplaceAllString(s, "${1}i${2}")
				fixes++
			}

			// 5. Espaçamento incorreto antes de pontuação
			if ocrPunctSpacing.MatchString(s) {
				s = ocrPunctSpacing.ReplaceAllString(s, "$1")
				fixes++
			}

			// 6. Espaços duplos redundantes
			if ocrDoubleSpaces.MatchString(s) {
				s = ocrDoubleSpaces.ReplaceAllString(s, " ")
				if len(s) != origLen {
					fixes++
				}
			}

			out.WriteString(s)
		}
	}

	return out.Bytes(), fixes
}

// FixEPUB repara problemas estruturais em arquivos EPUB para compatibilidade com Amazon Send to Kindle.
func FixEPUB(epubPath string) (int, error) {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return 0, fmt.Errorf("falha ao abrir arquivo EPUB: %w", err)
	}
	defer r.Close()

	tmpFile, err := os.CreateTemp(filepath.Dir(epubPath), "fixed-epub-*.epub")
	if err != nil {
		return 0, fmt.Errorf("falha ao criar arquivo temporário: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName) // Limpa em caso de erro

	zw := zip.NewWriter(tmpFile)
	fixesCount := 0

	// 1. O mimetype precisa ser o primeiro ficheiro e sem compressão (Store)
	var mimetypeHeader *zip.FileHeader
	var mimetypeContent []byte
	for _, f := range r.File {
		if f.Name == "mimetype" {
			rc, err := f.Open()
			if err == nil {
				mimetypeContent, _ = io.ReadAll(rc)
				rc.Close()
			}
			mimetypeHeader = &zip.FileHeader{
				Name:   "mimetype",
				Method: zip.Store,
			}
			break
		}
	}
	if mimetypeHeader == nil {
		mimetypeHeader = &zip.FileHeader{
			Name:   "mimetype",
			Method: zip.Store,
		}
		mimetypeContent = []byte("application/epub+zip")
		fixesCount++
	}

	w, err := zw.CreateHeader(mimetypeHeader)
	if err != nil {
		_ = zw.Close()
		_ = tmpFile.Close()
		return 0, err
	}
	_, _ = w.Write(bytes.TrimSpace(mimetypeContent))

	// 2. Itera sobre os outros ficheiros e aplica correções onde necessário
	for _, f := range r.File {
		if f.Name == "mimetype" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}

		lowerName := strings.ToLower(f.Name)
		modified := false

		// Correções em ficheiros OPF
		if strings.HasSuffix(lowerName, ".opf") {
			// Normaliza o idioma se estiver ausente ou incorreto
			if langMatch := langTagRegex.FindSubmatch(content); len(langMatch) > 1 {
				lang := strings.TrimSpace(string(langMatch[1]))
				if lang == "" || strings.EqualFold(lang, "und") || len(lang) > 12 {
					content = langTagRegex.ReplaceAll(content, []byte("<dc:language>en</dc:language>"))
					fixesCount++
					modified = true
				}
			} else {
				// Adiciona dc:language caso não exista
				if bytes.Contains(content, []byte("</metadata>")) {
					content = bytes.Replace(content, []byte("</metadata>"), []byte("  <dc:language>en</dc:language>\n  </metadata>"), 1)
					fixesCount++
					modified = true
				}
			}
		}

		// Correções em XHTML / HTML (tags não fechadas que quebram o parser do Kindle)
		if strings.HasSuffix(lowerName, ".xhtml") || strings.HasSuffix(lowerName, ".html") || strings.HasSuffix(lowerName, ".htm") {
			// Fecha tags <img> soltas
			if strayImgRegex.Match(content) {
				content = strayImgRegex.ReplaceAll(content, []byte("<img$1 />"))
				fixesCount++
				modified = true
			}
			// Fecha tags <br> soltas
			if strayBrRegex.Match(content) {
				content = strayBrRegex.ReplaceAll(content, []byte("<br$1/>"))
				fixesCount++
				modified = true
			}

			// Corrige declarações XML duplicadas no início do ficheiro
			decls := xmlDeclRegex.FindAll(content, -1)
			if len(decls) > 1 {
				first := decls[0]
				content = xmlDeclRegex.ReplaceAll(content, nil)
				content = append(first, append([]byte("\n"), content...)...)
				fixesCount++
				modified = true
			}

			// Aplica limpeza de artefactos e erros de digitalização OCR
			if cleaned, ocrFixes := CleanOCRErrors(content); ocrFixes > 0 {
				content = cleaned
				fixesCount += ocrFixes
				modified = true
			}
		}

		fh := &zip.FileHeader{
			Name:   f.Name,
			Method: zip.Deflate,
		}
		if !modified {
			fh.Method = f.Method
		}

		w, err := zw.CreateHeader(fh)
		if err != nil {
			_ = zw.Close()
			_ = tmpFile.Close()
			return 0, err
		}
		if _, err := w.Write(content); err != nil {
			_ = zw.Close()
			_ = tmpFile.Close()
			return 0, err
		}
	}

	if err := zw.Close(); err != nil {
		_ = tmpFile.Close()
		return 0, err
	}
	if err := tmpFile.Close(); err != nil {
		return 0, err
	}

	// Substituição atómica do arquivo original pelo corrigido
	r.Close() // Fecha o reader antes de mover no Windows
	if err := os.Rename(tmpName, epubPath); err != nil {
		// Fallback para cópia se rename falhar entre partições
		data, rErr := os.ReadFile(tmpName)
		if rErr != nil {
			return 0, err
		}
		if wErr := os.WriteFile(epubPath, data, 0644); wErr != nil {
			return 0, wErr
		}
	}

	return fixesCount, nil
}

