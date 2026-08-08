package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestLoadFromReturnsNotExistError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yml")

	_, err := LoadFrom(path)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadFrom() error = %v, want %v", err, os.ErrNotExist)
	}
}

func TestDefaultPathUsesOvermindConfigDirectory(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}

	wantSuffix := filepath.Join("overmind", "config.yml")
	if got := path; !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("DefaultPath() = %q, want suffix %q", got, wantSuffix)
	}
}

func TestLoadFromMergesFileWithDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	err := os.WriteFile(path, []byte(`vault:
  localRoot: ./notes
logging:
  level: debug
`), 0o644)
	if err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if cfg.Vault.Driver != "local" {
		t.Fatalf("Vault.Driver = %q, want %q", cfg.Vault.Driver, "local")
	}
	if cfg.Vault.RootPath != "./notes" {
		t.Fatalf("Vault.RootPath = %q, want %q", cfg.Vault.RootPath, "./notes")
	}
	if cfg.Logging.Level != "debug" {
		t.Fatalf("Logging.Level = %q, want %q", cfg.Logging.Level, "debug")
	}
}

func TestLoadFromRequiresRootPathForLocalDriver(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	err := os.WriteFile(path, []byte(`vault:
  driver: local
logging:
  level: info
`), 0o644)
	if err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	_, err = LoadFrom(path)
	if err == nil || !strings.Contains(err.Error(), "vault.localRoot is required") {
		t.Fatalf("LoadFrom() error = %v, want missing vault.localRoot error", err)
	}
}

func TestLoadFromRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	err := os.WriteFile(path, []byte(`vault:
  driver: local
logging:
  levle: debug
`), 0o644)
	if err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if _, err := LoadFrom(path); err == nil {
		t.Fatal("LoadFrom() error = nil, want unknown field error")
	}
}

func TestLoadFromRejectsMultipleDocuments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	err := os.WriteFile(path, []byte(`vault:
  driver: local
---
logging:
  level: debug
`), 0o644)
	if err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if _, err := LoadFrom(path); err == nil {
		t.Fatal("LoadFrom() error = nil, want multiple document error")
	}
}

func TestWriteDefaultCreatesConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "overmind", "config.yml")

	if err := WriteDefault(path, false); err != nil {
		t.Fatalf("WriteDefault() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	var cfg Config
	if err := yaml.Load(content, &cfg); err != nil {
		t.Fatalf("yaml.Load() error = %v", err)
	}
	if cfg.Vault.Driver != "local" || cfg.Vault.RootPath != "" || cfg.Logging.Level != "info" {
		t.Fatalf("written config = %+v, want local driver, empty localRoot and info logging", cfg)
	}
}
