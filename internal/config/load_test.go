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
	path, err := defaultPath()
	if err != nil {
		t.Fatalf("defaultPath() error = %v", err)
	}

	wantSuffix := filepath.Join("overmind", "config.yml")
	if got := path; !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("defaultPath() = %q, want suffix %q", got, wantSuffix)
	}
}

func TestLoadFromMergesFileWithDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	err := os.WriteFile(path, []byte(`vault:
  localRoot: ./notes
logging:
  level: debug
http:
  address: 0.0.0.0:9090
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
	if cfg.HTTP.Address != "0.0.0.0:9090" {
		t.Fatalf("HTTP.Address = %q, want %q", cfg.HTTP.Address, "0.0.0.0:9090")
	}
	if cfg.CLI.Mode != CLIModeLocal || cfg.CLI.Endpoint != "" {
		t.Fatalf("CLI = %+v, want local mode without endpoint", cfg.CLI)
	}
	if cfg.Index.Driver != "sqlite" || cfg.Index.Path != "./.overmind/index.db" {
		t.Fatalf("Index = %+v, want default SQLite index", cfg.Index)
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

func TestLoadFromRejectsUnsupportedVaultDriver(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	err := os.WriteFile(path, []byte(`vault:
  driver: remote
  localRoot: ./notes
logging:
  level: info
`), 0o644)
	if err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	_, err = LoadFrom(path)
	if err == nil || !strings.Contains(err.Error(), `unsupported vault.driver "remote"`) {
		t.Fatalf("LoadFrom() error = %v, want unsupported driver error", err)
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
	if cfg.Vault.Driver != "local" || cfg.Vault.RootPath != "" || cfg.Logging.Level != "info" || cfg.HTTP.Address != "127.0.0.1:8080" || cfg.CLI.Mode != CLIModeLocal || cfg.CLI.Endpoint != "" || cfg.Index.Driver != "sqlite" || cfg.Index.Path != "./.overmind/index.db" {
		t.Fatalf("written config = %+v, want default configuration", cfg)
	}
	if strings.Contains(string(content), "endpoint:") {
		t.Fatalf("default local config contains remote endpoint: %s", content)
	}
}

func TestLoadFromRejectsInvalidHTTPAddress(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	err := os.WriteFile(path, []byte(`vault:
  localRoot: ./notes
http:
  address: localhost
`), 0o644)
	if err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if _, err := LoadFrom(path); err == nil || !strings.Contains(err.Error(), "http.address") {
		t.Fatalf("LoadFrom() error = %v, want invalid http.address error", err)
	}
}

func TestLoadFromAllowsRemoteModeWithoutVault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	err := os.WriteFile(path, []byte(`cli:
  mode: remote
  endpoint: https://overmind.example/api
`), 0o644)
	if err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if cfg.CLI.Mode != CLIModeRemote || cfg.CLI.Endpoint != "https://overmind.example/api" {
		t.Fatalf("CLI = %+v", cfg.CLI)
	}
}

func TestLoadFromRejectsInvalidRemoteEndpoint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	err := os.WriteFile(path, []byte(`cli:
  mode: remote
  endpoint: overmind.example
`), 0o644)
	if err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if _, err := LoadFrom(path); err == nil || !strings.Contains(err.Error(), "cli.endpoint") {
		t.Fatalf("LoadFrom() error = %v, want invalid cli.endpoint error", err)
	}
}

func TestLoadFromRejectsUnsupportedLocalIndex(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	err := os.WriteFile(path, []byte(`vault:
  localRoot: ./notes
index:
  driver: json
`), 0o644)
	if err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if _, err := LoadFrom(path); err == nil || !strings.Contains(err.Error(), "index.driver") {
		t.Fatalf("LoadFrom() error = %v, want invalid index.driver error", err)
	}
}
