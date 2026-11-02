package cleaner

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var allowedBookExts = map[string]bool{
	".epub": true,
	".pdf":  true,
	".mobi": true,
	".azw":  true,
	".azw3": true,
	".cbz":  true,
	".cbr":  true,
}

// UnpackTransientArchive descompacta .zip ou .rar num diretório temporário isolado,
// filtra apenas os ficheiros de e-book suportados e descarta lixo promocional (.txt, .url, etc.).
func UnpackTransientArchive(archivePath, destinationIngestDir string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(archivePath))
	if ext != ".zip" && ext != ".rar" {
		return nil, fmt.Errorf("formato compactado não suportado: %s", ext)
	}

	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("cwa_transient_%s", uuid.New().String()[:8]))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("falha ao criar pasta temporária de descompressão: %w", err)
	}
	defer os.RemoveAll(tempDir) // Garante limpeza imediata do lixo após extração

	var extractedFiles []string
	var err error

	if ext == ".zip" {
		extractedFiles, err = extractZipSafe(archivePath, tempDir)
	} else if ext == ".rar" {
		extractedFiles, err = extractRarSafe(archivePath, tempDir)
	}

	if err != nil {
		return nil, fmt.Errorf("falha na extração de %s: %w", filepath.Base(archivePath), err)
	}

	var validEbooks []string
	for _, fPath := range extractedFiles {
		fileExt := strings.ToLower(filepath.Ext(fPath))
		if allowedBookExts[fileExt] {
			baseName := filepath.Base(fPath)
			destPath := filepath.Join(destinationIngestDir, baseName)

			// Evita colisão se já existir
			if _, statErr := os.Stat(destPath); statErr == nil {
				destPath = filepath.Join(destinationIngestDir, fmt.Sprintf("%s_%s%s", strings.TrimSuffix(baseName, fileExt), uuid.New().String()[:6], fileExt))
			}

			// Move o ficheiro do diretório temporário para a pasta de ingestão
			if moveErr := os.Rename(fPath, destPath); moveErr == nil {
				validEbooks = append(validEbooks, destPath)
			} else {
				// Fallback de cópia se rename entre filesystems diferentes falhar
				if copyErr := copyFile(fPath, destPath); copyErr == nil {
					validEbooks = append(validEbooks, destPath)
				}
			}
		}
	}

	return validEbooks, nil
}

// extractZipSafe extrai ficheiros .zip com proteção rigorosa contra ZipSlip (Path Traversal).
func extractZipSafe(zipPath, targetDir string) ([]string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	cleanTarget := filepath.Clean(targetDir) + string(os.PathSeparator)
	var files []string

	for _, f := range r.File {
		destPath := filepath.Join(targetDir, f.Name)
		cleanDest := filepath.Clean(destPath)

		// Proteção ZipSlip
		if !strings.HasPrefix(cleanDest, cleanTarget) {
			continue // Ignora tentativas maliciosas de path traversal
		}

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(cleanDest, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(cleanDest), 0755); err != nil {
			continue
		}

		outFile, err := os.OpenFile(cleanDest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			continue
		}

		_, copyErr := io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if copyErr == nil {
			files = append(files, cleanDest)
		}
	}

	return files, nil
}

// extractRarSafe extrai ficheiros .rar utilizando unrar ou 7z caso disponíveis no sistema.
func extractRarSafe(rarPath, targetDir string) ([]string, error) {
	cmdName := ""
	var args []string

	if _, err := exec.LookPath("unrar"); err == nil {
		cmdName = "unrar"
		args = []string{"x", "-y", "-inul", rarPath, targetDir}
	} else if _, err := exec.LookPath("7z"); err == nil {
		cmdName = "7z"
		args = []string{"x", "-y", fmt.Sprintf("-o%s", targetDir), rarPath}
	} else {
		return nil, fmt.Errorf("utilitário unrar/7z não encontrado no sistema para descompactar .rar")
	}

	cmd := exec.Command(cmdName, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("%s failed: %w (output: %s)", cmdName, err, string(out))
	}

	var extracted []string
	_ = filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			extracted = append(extracted, path)
		}
		return nil
	})

	return extracted, nil
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

