package backup

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type BackupInfo struct {
	Filename  string    `json:"filename"`
	SizeBytes int64     `json:"size_bytes"`
	SizeHuman string    `json:"size_human"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateBackup cria um snapshot comprimido (.zip) das bases de dados (metadata.db, app.db)
// e eventuais ficheiros de configuração/processed_books, mantendo a política de retenção.
func CreateBackup(libraryDir, backupDir string) (string, error) {
	if backupDir == "" {
		backupDir = filepath.Join(libraryDir, "backups")
	}
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("falha ao criar pasta de backups: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_150405")
	backupName := fmt.Sprintf("cwa_backup_%s.zip", timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	tmpPath := backupPath + ".tmp"
	outFile, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("falha ao criar ficheiro temporário de backup: %w", err)
	}
	defer func() {
		_ = outFile.Close()
		_ = os.Remove(tmpPath)
	}()

	zw := zip.NewWriter(outFile)

	// Ficheiros essenciais para snapshot
	targets := []string{
		"metadata.db",
		"metadata.db-wal",
		"app.db",
		"app.db-wal",
		"gdrive.db",
	}

	filesIncluded := 0
	for _, target := range targets {
		src := filepath.Join(libraryDir, target)
		info, err := os.Stat(src)
		if err != nil || info.IsDir() {
			continue
		}

		inF, err := os.Open(src)
		if err != nil {
			continue
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			inF.Close()
			continue
		}
		header.Name = target
		header.Method = zip.Deflate

		w, err := zw.CreateHeader(header)
		if err != nil {
			inF.Close()
			continue
		}

		_, _ = io.Copy(w, inF)
		inF.Close()
		filesIncluded++
	}

	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("falha ao fechar compressor zip: %w", err)
	}
	if err := outFile.Close(); err != nil {
		return "", fmt.Errorf("falha ao gravar ficheiro de backup: %w", err)
	}

	if err := os.Rename(tmpPath, backupPath); err != nil {
		return "", fmt.Errorf("falha ao renomear backup: %w", err)
	}

	// Auto-prune: Mantém os 10 backups mais recentes para prevenir saturação do disco
	_ = PruneBackups(backupDir, 10)

	return backupName, nil
}

// ListBackups devolve todos os backups .zip existentes ordenados do mais recente ao mais antigo.
func ListBackups(backupDir string) ([]BackupInfo, error) {
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}

	var backups []BackupInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".zip") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Filename:  e.Name(),
			SizeBytes: info.Size(),
			SizeHuman: formatSize(info.Size()),
			CreatedAt: info.ModTime(),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// PruneBackups remove backups antigos mantendo no máximo keepCount.
func PruneBackups(backupDir string, keepCount int) error {
	backups, err := ListBackups(backupDir)
	if err != nil || len(backups) <= keepCount {
		return nil
	}

	for i := keepCount; i < len(backups); i++ {
		_ = os.Remove(filepath.Join(backupDir, backups[i].Filename))
	}
	return nil
}

func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

