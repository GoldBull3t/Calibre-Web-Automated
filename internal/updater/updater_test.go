package updater_test

import (
	"context"
	"testing"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/updater"
)

func TestUpdaterCacheAndCheck(t *testing.T) {
	u := updater.GetDefault()
	if u == nil {
		t.Fatal("Updater não pode ser nulo")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	info, err := u.CheckUpdate(ctx)
	if err != nil {
		t.Fatalf("CheckUpdate falhou inesperadamente: %v", err)
	}

	if info.CurrentVersion == "" {
		t.Fatal("CurrentVersion não pode estar vazia")
	}

	// Segundo check deve vir do cache instantaneamente
	start := time.Now()
	info2, err := u.CheckUpdate(ctx)
	if err != nil {
		t.Fatalf("Segundo check falhou: %v", err)
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Errorf("Segundo check devia ter sido servido pelo cache local em <50ms")
	}
	if info2.LatestVersion != info.LatestVersion {
		t.Errorf("Versões não coincidem entre cache")
	}
}

