package formats

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	titleRegex   = regexp.MustCompile(`(?i)<dc:title>([^<]*)</dc:title>`)
	creatorRegex = regexp.MustCompile(`(?i)<dc:creator(\s+[^>]*)?>([^<]*)</dc:creator>`)
	metaCover    = regexp.MustCompile(`(?i)<meta\s+name=["']cover["']\s+content=["']([^"']+)["']`)
	manifestTag  = regexp.MustCompile(`(?i)(<manifest[^>]*>)`)
	metadataTag  = regexp.MustCompile(`(?i)(<metadata[^>]*>)`)
)

// EnforceCoverAndMetadata aplica diretamente ao ficheiro EPUB a capa física e os metadados (Título, Autor).
// Isso garante o compromisso CWA "Automatic Enforcement of Covers & Metadata": o ficheiro físico é 100% fiel
// ao que o utilizador configurou no Calibre-Web, para que leitores como o Kindle exibam a capa e título corretos.
func EnforceCoverAndMetadata(epubPath, coverImagePath, newTitle, newAuthor string) error {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return fmt.Errorf("falha ao abrir ficheiro EPUB para enforcement: %w", err)
	}

	var newCoverData []byte
	if coverImagePath != "" {
		data, err := os.ReadFile(coverImagePath)
		if err == nil && len(data) > 0 {
			newCoverData = data
		}
	}

	// 1. Localiza o ficheiro OPF no EPUB
	opfPath := ""
	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".opf") {
			opfPath = f.Name
			break
		}
	}

	// 2. Lê e analisa o OPF para encontrar o caminho interno da capa
	var coverEntryPath string
	var opfContent []byte
	if opfPath != "" {
		for _, f := range r.File {
			if f.Name == opfPath {
				rc, err := f.Open()
				if err == nil {
					opfContent, _ = io.ReadAll(rc)
					rc.Close()
				}
				break
			}
		}
	}

	if len(opfContent) > 0 {
		coverID := ""
		if m := metaCover.FindSubmatch(opfContent); len(m) > 1 {
			coverID = string(m[1])
		}

		itemRegex := regexp.MustCompile(`(?i)<item\s+[^>]*>`)
		for _, itemMatch := range itemRegex.FindAll(opfContent, -1) {
			str := string(itemMatch)
			idMatch := regexp.MustCompile(`(?i)id=["']([^"']+)["']`).FindStringSubmatch(str)
			hrefMatch := regexp.MustCompile(`(?i)href=["']([^"']+)["']`).FindStringSubmatch(str)
			propMatch := strings.Contains(strings.ToLower(str), "cover-image")

			if len(hrefMatch) > 1 {
				href := hrefMatch[1]
				isCover := false
				if propMatch {
					isCover = true
				} else if coverID != "" && len(idMatch) > 1 && idMatch[1] == coverID {
					isCover = true
				} else if len(idMatch) > 1 && (idMatch[1] == "cover" || idMatch[1] == "cover-image") {
					isCover = true
				}

				if isCover {
					dir := path.Dir(opfPath)
					if dir == "." || dir == "/" {
						coverEntryPath = href
					} else {
						coverEntryPath = path.Join(dir, href)
					}
					break
				}
			}
		}
	}

	// Se não encontrou caminho existente mas temos nova capa, usaremos um padrão
	if coverEntryPath == "" && len(newCoverData) > 0 {
		dir := path.Dir(opfPath)
		if dir == "." || dir == "/" {
			coverEntryPath = "cover.jpg"
		} else {
			coverEntryPath = path.Join(dir, "cover.jpg")
		}
	}

	// 3. Atualiza os metadados dentro do OPF
	if len(opfContent) > 0 {
		if newTitle != "" {
			if titleRegex.Match(opfContent) {
				opfContent = titleRegex.ReplaceAll(opfContent, fmt.Appendf(nil, "<dc:title>%s</dc:title>", escapeXML(newTitle)))
			} else {
				opfContent = metadataTag.ReplaceAll(opfContent, fmt.Appendf(nil, "$1\n    <dc:title>%s</dc:title>", escapeXML(newTitle)))
			}
		}
		if newAuthor != "" {
			if creatorRegex.Match(opfContent) {
				opfContent = creatorRegex.ReplaceAll(opfContent, fmt.Appendf(nil, "<dc:creator>%s</dc:creator>", escapeXML(newAuthor)))
			} else {
				opfContent = metadataTag.ReplaceAll(opfContent, fmt.Appendf(nil, "$1\n    <dc:creator>%s</dc:creator>", escapeXML(newAuthor)))
			}
		}
		// Se injetamos uma nova entrada de capa que não existia no OPF
		if len(newCoverData) > 0 && !strings.Contains(string(opfContent), coverEntryPath) {
			relHref := coverEntryPath
			if dir := path.Dir(opfPath); dir != "." && dir != "/" {
				relHref = strings.TrimPrefix(coverEntryPath, dir+"/")
			}
			metaItem := fmt.Sprintf(`$1`+"\n"+`    <item id="cwa-cover" href="%s" media-type="image/jpeg" properties="cover-image"/>`, relHref)
			opfContent = manifestTag.ReplaceAll(opfContent, []byte(metaItem))
			if !metaCover.Match(opfContent) {
				opfContent = metadataTag.ReplaceAll(opfContent, []byte(`$1`+"\n"+`    <meta name="cover" content="cwa-cover"/>`))
			}
		}
	}

	// 4. Cria ficheiro temporário para reescrever o ZIP
	tmpFile, err := os.CreateTemp(filepath.Dir(epubPath), "enforced-epub-*.epub")
	if err != nil {
		return fmt.Errorf("falha ao criar temporário: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	zw := zip.NewWriter(tmpFile)

	// Primeiro ficheiro: mimetype sem compressão
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
	}

	w, err := zw.CreateHeader(mimetypeHeader)
	if err != nil {
		_ = zw.Close()
		_ = tmpFile.Close()
		return err
	}
	_, _ = w.Write(bytes.TrimSpace(mimetypeContent))

	// Reescreve todas as outras entradas
	coverWritten := false
	for _, f := range r.File {
		if f.Name == "mimetype" {
			continue
		}

		if f.Name == opfPath && len(opfContent) > 0 {
			hdr := &zip.FileHeader{
				Name:   f.Name,
				Method: zip.Deflate,
			}
			ow, err := zw.CreateHeader(hdr)
			if err == nil {
				_, _ = ow.Write(opfContent)
			}
			continue
		}

		if coverEntryPath != "" && f.Name == coverEntryPath && len(newCoverData) > 0 {
			hdr := &zip.FileHeader{
				Name:   f.Name,
				Method: zip.Deflate,
			}
			cw, err := zw.CreateHeader(hdr)
			if err == nil {
				_, _ = cw.Write(newCoverData)
				coverWritten = true
			}
			continue
		}

		// Copia arquivo original
		rc, err := f.Open()
		if err != nil {
			continue
		}
		hdr := &zip.FileHeader{
			Name:   f.Name,
			Method: f.Method,
		}
		fw, err := zw.CreateHeader(hdr)
		if err == nil {
			_, _ = io.Copy(fw, rc)
		}
		rc.Close()
	}

	// Se a capa ainda não foi escrita (nova entrada criada)
	if !coverWritten && len(newCoverData) > 0 && coverEntryPath != "" {
		hdr := &zip.FileHeader{
			Name:   coverEntryPath,
			Method: zip.Deflate,
		}
		cw, err := zw.CreateHeader(hdr)
		if err == nil {
			_, _ = cw.Write(newCoverData)
		}
	}

	if err := zw.Close(); err != nil {
		_ = tmpFile.Close()
		_ = r.Close()
		return err
	}
	_ = tmpFile.Close()
	_ = r.Close()

	// Substituição atómica do ficheiro
	return os.Rename(tmpName, epubPath)
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
