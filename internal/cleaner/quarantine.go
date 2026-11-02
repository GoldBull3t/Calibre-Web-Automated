package cleaner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/database"
	"github.com/google/uuid"
)

// QuarantineFile isola ficheiros corrompidos ou com parsing ambíguo na pasta de quarentena
// e regista o evento na base de dados para exposição na interface e resolução manual.
func QuarantineFile(ctx context.Context, filePath, quarantineDir, reason string, errDetails error, db *database.DB) error {
	if quarantineDir == "" {
		quarantineDir = "/cwa-quarantine"
	}

	hash, size, hashErr := CalculateSHA256(filePath)
	if hashErr != nil || hash == "" {
		hash = uuid.New().String()
	}

	folderID := hash
	if len(folderID) > 16 {
		folderID = folderID[:16]
	}

	targetFolder := filepath.Join(quarantineDir, folderID)
	if err := os.MkdirAll(targetFolder, 0755); err != nil {
		return fmt.Errorf("falha ao criar pasta de quarentena: %w", err)
	}

	filename := filepath.Base(filePath)
	targetPath := filepath.Join(targetFolder, filename)

	// Move o ficheiro
	if err := os.Rename(filePath, targetPath); err != nil {
		if copyErr := copyFile(filePath, targetPath); copyErr != nil {
			return fmt.Errorf("falha ao mover ficheiro para quarentena: %w", copyErr)
		}
		_ = os.Remove(filePath)
	}

	errStr := ""
	if errDetails != nil {
		errStr = errDetails.Error()
	}

	// Regista o evento na base de dados
	if db != nil {
		item := database.QuarantineItem{
			ID:           uuid.New().String(),
			Filename:     filename,
			FilePath:     targetPath,
			FileHash:     hash,
			FileSize:     size,
			Reason:       reason,
			ErrorDetails: errStr,
			Status:       "pending",
			CreatedAt:    time.Now().UTC(),
		}
		_ = db.AddQuarantineItem(ctx, item)
	}

	return nil
}

