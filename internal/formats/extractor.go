package formats

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/crocodilestick/calibre-web-go/internal/epub"
)

type ExtractedBook struct {
	Title       string
	Authors     []string
	Description string
	Publisher   string
	Language    string
	Identifiers map[string]string
	CoverBytes  []byte
	Format      string
}

// Extract analisa o ficheiro fornecido de acordo com a extensão e extrai metadados e capa.
func Extract(filePath string) (*ExtractedBook, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".epub":
		return extractEPUB(filePath)
	case ".cbz":
		return extractCBZ(filePath)
	case ".mobi", ".azw", ".azw3":
		return extractMOBI(filePath)
	case ".pdf":
		return extractPDF(filePath)
	case ".fb2":
		return extractFB2(filePath)
	case ".txt":
		return extractTXT(filePath)
	default:
		return extractGeneric(filePath, strings.ToUpper(strings.TrimPrefix(ext, ".")))
	}
}

func extractEPUB(filePath string) (*ExtractedBook, error) {
	meta, err := epub.ExtractMetadata(filePath)
	if err != nil {
		return nil, err
	}

	book := &ExtractedBook{
		Title:       meta.Title,
		Authors:     meta.Authors,
		Description: meta.Description,
		Publisher:   meta.Publisher,
		Language:    meta.Language,
		Identifiers: meta.Identifiers,
		Format:      "EPUB",
	}

	var coverBuf bytes.Buffer
	if err := epub.ExtractCover(filePath, &coverBuf); err == nil && coverBuf.Len() > 0 {
		book.CoverBytes = coverBuf.Bytes()
	}

	return book, nil
}

type comicInfoXML struct {
	Title   string `xml:"Title"`
	Series  string `xml:"Series"`
	Writer  string `xml:"Writer"`
	Summary string `xml:"Summary"`
}

func extractCBZ(filePath string) (*ExtractedBook, error) {
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir CBZ: %w", err)
	}
	defer zr.Close()

	book := &ExtractedBook{
		Format:      "CBZ",
		Identifiers: make(map[string]string),
	}

	var imageFiles []string
	fileMap := make(map[string]*zip.File)

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		nameLower := strings.ToLower(f.Name)
		fileMap[f.Name] = f

		if nameLower == "comicinfo.xml" {
			rc, err := f.Open()
			if err == nil {
				var ci comicInfoXML
				if xml.NewDecoder(rc).Decode(&ci) == nil {
					book.Title = ci.Title
					if ci.Writer != "" {
						book.Authors = []string{ci.Writer}
					}
					book.Description = ci.Summary
				}
				rc.Close()
			}
		}

		if strings.HasSuffix(nameLower, ".jpg") || strings.HasSuffix(nameLower, ".jpeg") ||
			strings.HasSuffix(nameLower, ".png") || strings.HasSuffix(nameLower, ".webp") {
			imageFiles = append(imageFiles, f.Name)
		}
	}

	sort.Strings(imageFiles)
	if len(imageFiles) > 0 {
		firstImg := fileMap[imageFiles[0]]
		if rc, err := firstImg.Open(); err == nil {
			data, _ := io.ReadAll(rc)
			rc.Close()
			book.CoverBytes = data
		}
	}

	if book.Title == "" {
		book.Title, book.Authors = parseFilename(filePath)
	}

	return book, nil
}

func extractMOBI(filePath string) (*ExtractedBook, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(filePath), "."))
	book := &ExtractedBook{
		Format:      ext,
		Identifiers: make(map[string]string),
	}

	// Lê cabeçalho Palm Database (78 bytes iniciais)
	header := make([]byte, 78)
	if _, err := io.ReadFull(f, header); err != nil {
		book.Title, book.Authors = parseFilename(filePath)
		return book, nil
	}

	numRecords := binary.BigEndian.Uint16(header[76:78])
	if numRecords == 0 {
		book.Title, book.Authors = parseFilename(filePath)
		return book, nil
	}

	recordInfoList := make([]byte, numRecords*8)
	if _, err := io.ReadFull(f, recordInfoList); err != nil {
		book.Title, book.Authors = parseFilename(filePath)
		return book, nil
	}

	record0Offset := binary.BigEndian.Uint32(recordInfoList[0:4])
	var record1Offset uint32
	if numRecords > 1 {
		record1Offset = binary.BigEndian.Uint32(recordInfoList[8:12])
	}

	if record1Offset > record0Offset && record1Offset-record0Offset > 16 {
		rec0 := make([]byte, record1Offset-record0Offset)
		if _, err := f.Seek(int64(record0Offset), io.SeekStart); err == nil {
			if _, err := io.ReadFull(f, rec0); err == nil {
				// Verifica Magic "MOBI" no offset 16 do PalmDOC record 0
				if len(rec0) >= 24 && string(rec0[16:20]) == "MOBI" {
					if len(rec0) >= 92 {
						fullNameOffset := binary.BigEndian.Uint32(rec0[84:88])
						fullNameLen := binary.BigEndian.Uint32(rec0[88:92])
						if int(fullNameOffset+fullNameLen) <= len(rec0) {
							book.Title = string(rec0[fullNameOffset : fullNameOffset+fullNameLen])
						}
					}

					// Verifica bloco EXTH
					exthFlags := binary.BigEndian.Uint32(rec0[128:132])
					if (exthFlags & 0x40) != 0 {
						mobiHeaderLen := binary.BigEndian.Uint32(rec0[20:24])
						exthOffset := 16 + mobiHeaderLen
						if int(exthOffset+12) <= len(rec0) && string(rec0[exthOffset:exthOffset+4]) == "EXTH" {
							recCount := binary.BigEndian.Uint32(rec0[exthOffset+8 : exthOffset+12])
							pos := exthOffset + 12
							for i := uint32(0); i < recCount; i++ {
								if int(pos+8) > len(rec0) {
									break
								}
								rType := binary.BigEndian.Uint32(rec0[pos : pos+4])
								rLen := binary.BigEndian.Uint32(rec0[pos+4 : pos+8])
								if rLen < 8 || int(pos+rLen) > len(rec0) {
									break
								}
								val := string(rec0[pos+8 : pos+rLen])
								switch rType {
								case 100: // Author
									book.Authors = append(book.Authors, strings.TrimSpace(val))
								case 101: // Publisher
									book.Publisher = strings.TrimSpace(val)
								case 103: // Description
									book.Description = strings.TrimSpace(val)
								case 524: // Language
									book.Language = strings.TrimSpace(val)
								}
								pos += rLen
							}
						}
					}
				}
			}
		}
	}

	if book.Title == "" {
		book.Title, book.Authors = parseFilename(filePath)
	}

	return book, nil
}

func extractPDF(filePath string) (*ExtractedBook, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	book := &ExtractedBook{
		Format:      "PDF",
		Identifiers: make(map[string]string),
	}

	// Lê os últimos 8KB do PDF onde geralmente reside o dicionário /Info
	stat, err := f.Stat()
	if err == nil && stat.Size() > 0 {
		readSize := int64(8192)
		if readSize > stat.Size() {
			readSize = stat.Size()
		}
		buf := make([]byte, readSize)
		if _, err := f.Seek(stat.Size()-readSize, io.SeekStart); err == nil {
			if _, err := io.ReadFull(f, buf); err == nil {
				str := string(buf)
				if idx := strings.Index(str, "/Title ("); idx != -1 {
					sub := str[idx+8:]
					if end := strings.Index(sub, ")"); end != -1 {
						book.Title = strings.TrimSpace(sub[:end])
					}
				}
				if idx := strings.Index(str, "/Author ("); idx != -1 {
					sub := str[idx+9:]
					if end := strings.Index(sub, ")"); end != -1 {
						book.Authors = []string{strings.TrimSpace(sub[:end])}
					}
				}
			}
		}
	}

	if book.Title == "" {
		book.Title, book.Authors = parseFilename(filePath)
	}

	return book, nil
}

func extractFB2(filePath string) (*ExtractedBook, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	type fb2TitleInfo struct {
		BookTitle string `xml:"title-info>book-title"`
		Authors   []struct {
			FirstName string `xml:"first-name"`
			LastName  string `xml:"last-name"`
		} `xml:"title-info>author"`
		Annotation string `xml:"title-info>annotation"`
	}

	type fb2Doc struct {
		XMLName     xml.Name     `xml:"FictionBook"`
		Description fb2TitleInfo `xml:"description"`
	}

	var doc fb2Doc
	if err := xml.NewDecoder(f).Decode(&doc); err == nil {
		book := &ExtractedBook{
			Title:       strings.TrimSpace(doc.Description.BookTitle),
			Description: strings.TrimSpace(doc.Description.Annotation),
			Format:      "FB2",
			Identifiers: make(map[string]string),
		}
		for _, a := range doc.Description.Authors {
			name := strings.TrimSpace(fmt.Sprintf("%s %s", a.FirstName, a.LastName))
			if name != "" {
				book.Authors = append(book.Authors, name)
			}
		}
		if book.Title != "" {
			return book, nil
		}
	}

	return extractGeneric(filePath, "FB2")
}

func extractTXT(filePath string) (*ExtractedBook, error) {
	title, authors := parseFilename(filePath)
	return &ExtractedBook{
		Title:       title,
		Authors:     authors,
		Format:      "TXT",
		Identifiers: make(map[string]string),
	}, nil
}

func extractGeneric(filePath, format string) (*ExtractedBook, error) {
	title, authors := parseFilename(filePath)
	return &ExtractedBook{
		Title:       title,
		Authors:     authors,
		Format:      format,
		Identifiers: make(map[string]string),
	}, nil
}

func parseFilename(filePath string) (title string, authors []string) {
	base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	parts := strings.Split(base, " - ")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[0]), []string{strings.TrimSpace(parts[1])}
	}
	return strings.TrimSpace(base), []string{"Unknown"}
}

// UnpackZip extrai ficheiros de ebook suportados a partir de um ficheiro ZIP.
func UnpackZip(zipPath, destDir string) ([]string, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao descompactar zip: %w", err)
	}
	defer zr.Close()

	supportedExts := map[string]bool{
		".epub": true, ".pdf": true, ".mobi": true, ".azw": true,
		".azw3": true, ".cbz": true, ".txt": true, ".fb2": true,
	}

	var extractedFiles []string

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name))
		if !supportedExts[ext] {
			continue
		}

		baseName := filepath.Base(f.Name)
		destPath := filepath.Join(destDir, baseName)

		rc, err := f.Open()
		if err != nil {
			continue
		}

		out, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			continue
		}

		_, _ = io.Copy(out, rc)
		out.Close()
		rc.Close()

		extractedFiles = append(extractedFiles, destPath)
	}

	return extractedFiles, nil
}

