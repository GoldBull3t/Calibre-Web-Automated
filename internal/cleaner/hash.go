package cleaner

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// CalculateSHA256 calcula o hash SHA-256 e o tamanho de um ficheiro em streaming (buffer de 64KB).
func CalculateSHA256(filePath string) (string, int64, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("falha ao abrir ficheiro para hash: %w", err)
	}
	defer f.Close()

	hasher := sha256.New()
	buf := make([]byte, 64*1024)
	n, err := io.CopyBuffer(hasher, f, buf)
	if err != nil {
		return "", 0, fmt.Errorf("falha ao calcular SHA-256: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), n, nil
}
