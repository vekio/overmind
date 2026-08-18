package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigurationFilesUseIndependentDirectories(t *testing.T) {
	configHome := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "data"))

	cliFile, _, err := NewCLIFile()
	if err != nil {
		t.Fatalf("NewCLIFile() error = %v", err)
	}
	serverFile, _, err := NewServerFile()
	if err != nil {
		t.Fatalf("NewServerFile() error = %v", err)
	}

	if want := filepath.Join(configHome, "overmind", "config.yml"); cliFile.Path() != want {
		t.Fatalf("CLI path = %q, want %q", cliFile.Path(), want)
	}
	if want := filepath.Join(configHome, "overmind-server", "config.yml"); serverFile.Path() != want {
		t.Fatalf("server path = %q, want %q", serverFile.Path(), want)
	}
}

func TestConfigurationDefaultsShareApplicationData(t *testing.T) {
	dataHome := filepath.Join(t.TempDir(), "data")
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))
	t.Setenv("XDG_DATA_HOME", dataHome)

	_, cliDefaults, err := NewCLIFile()
	if err != nil {
		t.Fatalf("NewCLIFile() error = %v", err)
	}
	_, serverDefaults, err := NewServerFile()
	if err != nil {
		t.Fatalf("NewServerFile() error = %v", err)
	}

	wantVault := filepath.Join(dataHome, "overmind", "vault")
	wantIndex := filepath.Join(dataHome, "overmind", "index.db")
	if cliDefaults.Vault.RootPath != wantVault || serverDefaults.Vault.RootPath != wantVault {
		t.Fatalf("vault paths = %q, %q; want %q", cliDefaults.Vault.RootPath, serverDefaults.Vault.RootPath, wantVault)
	}
	if cliDefaults.Index.Path != wantIndex || serverDefaults.Index.Path != wantIndex {
		t.Fatalf("index paths = %q, %q; want %q", cliDefaults.Index.Path, serverDefaults.Index.Path, wantIndex)
	}
	if serverDefaults.HTTP.Address != "127.0.0.1:8080" {
		t.Fatalf("server HTTP address = %q", serverDefaults.HTTP.Address)
	}
}

func TestCLILoadOrCreateWritesPrivateCompleteConfiguration(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "data"))
	file, defaults, err := NewCLIFile()
	if err != nil {
		t.Fatalf("NewCLIFile() error = %v", err)
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

func TestConfigurationFilesRejectFieldsFromTheOtherProcess(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "data"))

	cliFile, _, err := NewCLIFile()
	if err != nil {
		t.Fatal(err)
	}
	writeConfigContent(t, cliFile.Path(), `vault:
  driver: local
  localRoot: ./vault
logging:
  level: info
index:
  driver: sqlite
  path: ./index.db
http:
  address: 127.0.0.1:8080
`)
	if _, err := cliFile.Load(); err == nil || !strings.Contains(err.Error(), "http") {
		t.Fatalf("CLI Load() error = %v, want unknown HTTP field", err)
	}

	serverFile, _, err := NewServerFile()
	if err != nil {
		t.Fatal(err)
	}
	writeConfigContent(t, serverFile.Path(), `vault:
  driver: local
  localRoot: ./vault
logging:
  level: info
index:
  driver: sqlite
  path: ./index.db
cli:
  mode: local
http:
  address: 127.0.0.1:8080
`)
	if _, err := serverFile.Load(); err == nil || !strings.Contains(err.Error(), "cli") {
		t.Fatalf("server Load() error = %v, want unknown CLI field", err)
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
