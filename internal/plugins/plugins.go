// Package plugins gere a instalação e ativação de plugins Calibre (.zip) guardados em disco.
package plugins

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const disabledSuffix = ".disabled"

// Plugin descreve um plugin instalado.
type Plugin struct {
	Filename string `json:"filename"`
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	Size     int64  `json:"size"`
}

// Manager gere os ficheiros de plugins num diretório.
type Manager struct {
	dir string
	mu  sync.Mutex
}

// NewManager cria o gestor e garante que o diretório existe.
func NewManager(dir string) (*Manager, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Manager{dir: dir}, nil
}

func baseName(filename string) string {
	return strings.TrimSuffix(filepath.Base(filename), disabledSuffix)
}

// List devolve os plugins instalados, ordenados por nome.
func (m *Manager) List() []Plugin {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []Plugin{}
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		enabled := strings.HasSuffix(strings.ToLower(n), ".zip")
		if !enabled && !strings.HasSuffix(strings.ToLower(n), ".zip"+disabledSuffix) {
			continue
		}
		var size int64
		if info, err := e.Info(); err == nil {
			size = info.Size()
		}
		base := baseName(n)
		out = append(out, Plugin{
			Filename: base,
			Name:     strings.TrimSuffix(base, filepath.Ext(base)),
			Enabled:  enabled,
			Size:     size,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// SavePlugin grava um plugin .zip no diretório de plugins.
func (m *Manager) SavePlugin(filename string, r io.Reader) (*Plugin, error) {
	name := baseName(filename)
	if !strings.HasSuffix(strings.ToLower(name), ".zip") {
		return nil, errors.New("o plugin tem de ser um ficheiro .zip")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	_ = os.Remove(filepath.Join(m.dir, name+disabledSuffix))
	f, err := os.Create(filepath.Join(m.dir, name))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	n, err := io.Copy(f, r)
	if err != nil {
		return nil, err
	}
	return &Plugin{Filename: name, Name: strings.TrimSuffix(name, filepath.Ext(name)), Enabled: true, Size: n}, nil
}

// Toggle ativa ou desativa um plugin existente.
func (m *Manager) Toggle(filename string, enabled bool) error {
	name := baseName(filename)
	m.mu.Lock()
	defer m.mu.Unlock()
	on := filepath.Join(m.dir, name)
	off := on + disabledSuffix
	src, dst := off, on
	if !enabled {
		src, dst = on, off
	}
	if _, err := os.Stat(src); err != nil {
		if _, err2 := os.Stat(dst); err2 == nil {
			return nil
		}
		return fmt.Errorf("plugin %q não encontrado", name)
	}
	return os.Rename(src, dst)
}

