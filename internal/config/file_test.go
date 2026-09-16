package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigurationFileUsesApplicationDirectory(t *testing.T) {
	configHome := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "data"))

	file, _, err := NewFile()
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}
	if want := filepath.Join(configHome, "overmind", "config.yml"); file.Path() != want {
		t.Fatalf("path = %q, want %q", file.Path(), want)
	}
}

func TestConfigurationDefaultsUseApplicationData(t *testing.T) {
	dataHome := filepath.Join(t.TempDir(), "data")
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))
	t.Setenv("XDG_DATA_HOME", dataHome)

	_, defaults, err := NewFile()
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}

	wantVault := filepath.Join(dataHome, "overmind", "vault")
	wantIndex := filepath.Join(dataHome, "overmind", "index.db")
	if defaults.Vault.RootPath != wantVault {
		t.Fatalf("vault path = %q, want %q", defaults.Vault.RootPath, wantVault)
	}
	if defaults.Index.Path != wantIndex {
		t.Fatalf("index path = %q, want %q", defaults.Index.Path, wantIndex)
	}
}

func TestLoadOrCreateWritesPrivateCompleteConfiguration(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "data"))
	file, defaults, err := NewFile()
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}

	loaded, err := file.LoadOrCreate(defaults)
	if err != nil {
		t.Fatalf("LoadOrCreate() error = %v", err)
	}
	if loaded != defaults {
		t.Fatalf("LoadOrCreate() = %+v, want %+v", loaded, defaults)
	}
	info, err := os.Stat(file.Path())
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("configuration permissions = %o, want 600", got)
	}
}

func TestConfigurationFileRejectsUnknownFields(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "data"))

	file, _, err := NewFile()
	if err != nil {
		t.Fatal(err)
	}
	writeConfigContent(t, file.Path(), `vault:
  driver: local
  localRoot: ./vault
index:
  driver: sqlite
  path: ./index.db
unexpected: true
`)
	if _, err := file.Load(); err == nil || !strings.Contains(err.Error(), "unexpected") {
		t.Fatalf("Load() error = %v, want unknown field", err)
	}
}

func writeConfigContent(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
