package backup_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/backup"
)

func TestBackupCreationAndPrune(t *testing.T) {
	tempLib := t.TempDir()
	tempBackups := filepath.Join(tempLib, "backups")

	// Cria ficheiros simulados
	_ = os.WriteFile(filepath.Join(tempLib, "metadata.db"), []byte("SQLITE_METADATA_DATA"), 0644)
	_ = os.WriteFile(filepath.Join(tempLib, "app.db"), []byte("SQLITE_APP_DATA"), 0644)

	name, err := backup.CreateBackup(tempLib, tempBackups)
	if err != nil {
		t.Fatalf("CreateBackup falhou: %v", err)
	}

	backups, err := backup.ListBackups(tempBackups)
	if err != nil {
		t.Fatalf("ListBackups falhou: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("Esperava 1 backup, obteve %d", len(backups))
	}
	if backups[0].Filename != name {
		t.Fatalf("Nome do backup diverge: %s vs %s", backups[0].Filename, name)
	}

	// Valida integridade do zip criado
	zr, err := zip.OpenReader(filepath.Join(tempBackups, name))
	if err != nil {
		t.Fatalf("Backup zip inválido: %v", err)
	}
	defer zr.Close()

	foundMetadata := false
	foundApp := false
	for _, f := range zr.File {
		if f.Name == "metadata.db" {
			foundMetadata = true
		}
		if f.Name == "app.db" {
			foundApp = true
		}
	}
	_ = zr.Close()

	if !foundMetadata || !foundApp {
		t.Fatalf("Ficheiros esperados ausentes no zip: metadata=%v, app=%v", foundMetadata, foundApp)
	}

	// Testa pruning
	time.Sleep(10 * time.Millisecond)
	_ = os.WriteFile(filepath.Join(tempBackups, "cwa_backup_old1.zip"), []byte("fake"), 0644)
	_ = os.WriteFile(filepath.Join(tempBackups, "cwa_backup_old2.zip"), []byte("fake"), 0644)

	_ = backup.PruneBackups(tempBackups, 2)
	updatedList, _ := backup.ListBackups(tempBackups)
	if len(updatedList) > 2 {
		t.Fatalf("PruneBackups devia ter reduzido a 2, tem %d", len(updatedList))
	}
}
