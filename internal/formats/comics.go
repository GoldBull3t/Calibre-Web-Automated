package formats

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// ComicInfo contém a lista de páginas de uma banda desenhada / mangá.
type ComicInfo struct {
	TotalPages int      `json:"total_pages"`
	Pages      []string `json:"pages"`
	IsRAR      bool     `json:"-"`
}

var (
	comicCacheMu sync.RWMutex
	comicCache   = make(map[string]*ComicInfo)
)

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

func isRARFile(filePath string) bool {
	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	var head [4]byte
	n, err := f.Read(head[:])
	return err == nil && n == 4 && head[0] == 0x52 && head[1] == 0x61 && head[2] == 0x72 && head[3] == 0x21
}

func getRARExtractor() (bin string, is7z bool, err error) {
	if bin, err := exec.LookPath("unrar"); err == nil {
		return bin, false, nil
	}
	if bin, err := exec.LookPath("7z"); err == nil {
		return bin, true, nil
	}
	return "", false, fmt.Errorf("nem 'unrar' nem '7z' (p7zip) estão instalados no sistema para processar CBR")
}

// ListComicPages devolve os caminhos internos de todas as imagens ordenadas de um CBZ/CBR.
func ListComicPages(filePath string) (*ComicInfo, error) {
	comicCacheMu.RLock()
	if cached, ok := comicCache[filePath]; ok {
		comicCacheMu.RUnlock()
		return cached, nil
	}
	comicCacheMu.RUnlock()

	var imagePaths []string
	isRAR := isRARFile(filePath)

	if isRAR {
		bin, is7z, err := getRARExtractor()
		if err != nil {
			return nil, err
		}

		var rawLines []string
		if is7z {
			out, err := exec.Command(bin, "l", "-ba", "-slt", "--", filePath).Output()
			if err != nil {
				return nil, fmt.Errorf("erro ao ler arquivo RAR via 7z: %w", err)
			}
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "Path = ") {
					rawLines = append(rawLines, strings.TrimPrefix(line, "Path = "))
				}
			}
		} else {
			out, err := exec.Command(bin, "lb", "--", filePath).Output()
			if err != nil {
				return nil, fmt.Errorf("erro ao ler arquivo RAR via unrar: %w", err)
			}
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					rawLines = append(rawLines, line)
				}
			}
		}

		for _, line := range rawLines {
			nameLower := strings.ToLower(line)
			if strings.HasPrefix(filepath.Base(nameLower), ".") || strings.Contains(nameLower, "__macosx") {
				continue
			}
			if strings.HasSuffix(nameLower, ".jpg") || strings.HasSuffix(nameLower, ".jpeg") ||
				strings.HasSuffix(nameLower, ".png") || strings.HasSuffix(nameLower, ".webp") ||
				strings.HasSuffix(nameLower, ".avif") || strings.HasSuffix(nameLower, ".gif") {
				imagePaths = append(imagePaths, line)
			}
		}
	} else {
		zr, err := zip.OpenReader(filePath)
		if err != nil {
			return nil, fmt.Errorf("erro ao abrir arquivo de banda desenhada: %w", err)
		}
		defer zr.Close()

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
	}

	sort.Slice(imagePaths, func(i, j int) bool {
		return naturalLess(imagePaths[i], imagePaths[j])
	})

	info := &ComicInfo{
		TotalPages: len(imagePaths),
		Pages:      imagePaths,
		IsRAR:      isRAR,
	}

	comicCacheMu.Lock()
	comicCache[filePath] = info
	comicCacheMu.Unlock()

	return info, nil
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

	if info.IsRAR {
		bin, is7z, err := getRARExtractor()
		if err != nil {
			return nil, "", err
		}
		var buf bytes.Buffer
		var cmd *exec.Cmd
		if is7z {
			cmd = exec.Command(bin, "e", "-so", "-bd", "--", filePath, targetName)
		} else {
			cmd = exec.Command(bin, "p", "-inul", "--", filePath, targetName)
		}
		cmd.Stdout = &buf
		if err := cmd.Run(); err != nil {
			return nil, "", fmt.Errorf("falha ao extrair página do CBR: %w", err)
		}
		return buf.Bytes(), mimeType, nil
	}

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

			return data, mimeType, nil
		}
	}

	return nil, "", fmt.Errorf("página não encontrada: %s", targetName)
}
