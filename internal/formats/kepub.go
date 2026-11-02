package formats

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerateKepub gera uma versão otimizada para Kobo (.kepub.epub) a partir de um EPUB padrão.
func GenerateKepub(epubPath, outputDir string) (string, error) {
	kepubifyPath := ""
	if p, err := exec.LookPath("kepubify"); err == nil {
		kepubifyPath = p
	} else if _, err := os.Stat("/usr/local/bin/kepubify"); err == nil {
		kepubifyPath = "/usr/local/bin/kepubify"
	}

	if kepubifyPath == "" {
		return "", fmt.Errorf("utilitário kepubify não encontrado no sistema")
	}

	baseName := filepath.Base(epubPath)
	cleanName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	outputFile := filepath.Join(outputDir, cleanName+".kepub.epub")

	cmd := exec.Command(kepubifyPath, "-o", outputFile, epubPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("kepubify falhou: %w (output: %s)", err, string(out))
	}

	if _, err := os.Stat(outputFile); err != nil {
		return "", fmt.Errorf("ficheiro kepub gerado não foi encontrado: %w", err)
	}

	return outputFile, nil
}
