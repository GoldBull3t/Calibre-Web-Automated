package formats

import (
	"archive/zip"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// ComicInfo contém a lista de páginas de uma banda desenhada / mangá.
type ComicInfo struct {
	TotalPages int      `json:"total_pages"`
	Pages      []string `json:"pages"`
}

// IsComicFormat verifica se o ficheiro tem extensão de banda desenhada / mangá.
func IsComicFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".cbz" || ext == ".cbr"
}

// naturalLess compara duas strings considerando sequências numéricas (Natural Sort).
func naturalLess(s1, s2 string) bool {
	chunks1 := chunkString(s1)
	chunks2 := chunkString(s2)

	n := len(chunks1)
	if len(chunks2) < n {
		n = len(chunks2)
	}

	for i := 0; i < n; i++ {
		c1, c2 := chunks1[i], chunks2[i]
		if isDigit(c1[0]) && isDigit(c2[0]) {
			num1, err1 := strconv.ParseUint(c1, 10, 64)
			num2, err2 := strconv.ParseUint(c2, 10, 64)
			if err1 == nil && err2 == nil && num1 != num2 {
				return num1 < num2
			}
		}
		if c1 != c2 {
			return c1 < c2
		}
	}
	return len(chunks1) < len(chunks2)
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func chunkString(s string) []string {
	var chunks []string
	runes := []rune(s)
	n := len(runes)
	i := 0
	for i < n {
		start := i
		isNum := unicode.IsDigit(runes[i])
		for i < n && unicode.IsDigit(runes[i]) == isNum {
			i++
		}
		chunks = append(chunks, string(runes[start:i]))
	}
	return chunks
}

// ListComicPages devolve os caminhos internos de todas as imagens ordenadas de um CBZ/CBR.
func ListComicPages(filePath string) (*ComicInfo, error) {
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo de banda desenhada: %w", err)
	}
	defer zr.Close()

	var imagePaths []string
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		nameLower := strings.ToLower(f.Name)
		if strings.HasPrefix(filepath.Base(nameLower), ".") || strings.Contains(nameLower, "__macosx") {
			continue
		}
		if strings.HasSuffix(nameLower, ".jpg") || strings.HasSuffix(nameLower, ".jpeg") ||
			strings.HasSuffix(nameLower, ".png") || strings.HasSuffix(nameLower, ".webp") ||
			strings.HasSuffix(nameLower, ".avif") || strings.HasSuffix(nameLower, ".gif") {
			imagePaths = append(imagePaths, f.Name)
		}
	}

	sort.Slice(imagePaths, func(i, j int) bool {
		return naturalLess(imagePaths[i], imagePaths[j])
	})

	return &ComicInfo{
		TotalPages: len(imagePaths),
		Pages:      imagePaths,
	}, nil
}

// ExtractComicPage extrai os bytes e o tipo MIME de uma página específica (por índice).
func ExtractComicPage(filePath string, pageIndex int) ([]byte, string, error) {
	info, err := ListComicPages(filePath)
	if err != nil {
		return nil, "", err
	}
	if pageIndex < 0 || pageIndex >= len(info.Pages) {
		return nil, "", fmt.Errorf("página %d fora do intervalo (total: %d)", pageIndex, len(info.Pages))
	}

	targetName := info.Pages[pageIndex]

	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, "", err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name == targetName {
			rc, err := f.Open()
			if err != nil {
				return nil, "", err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, "", err
			}

			mimeType := mime.TypeByExtension(filepath.Ext(targetName))
			if mimeType == "" {
				ext := strings.ToLower(filepath.Ext(targetName))
				switch ext {
				case ".jpg", ".jpeg":
					mimeType = "image/jpeg"
				case ".png":
					mimeType = "image/png"
				case ".webp":
					mimeType = "image/webp"
				case ".gif":
					mimeType = "image/gif"
				case ".avif":
					mimeType = "image/avif"
				default:
					mimeType = "application/octet-stream"
				}
			}

			return data, mimeType, nil
		}
	}

	return nil, "", fmt.Errorf("página não encontrada: %s", targetName)
}
