package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/vekio/overmind/internal/config"
)

func TestNormalizeVaultPathDoesNotCreateDirectories(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Chdir(root)
	for _, input := range []string{" ./new-vault ", "~/new-vault"} {
		path, err := config.NormalizeVaultPath(input)
		want := filepath.Join(root, "new-vault")
		if err != nil || path != want {
			t.Fatalf("NormalizeVaultPath(%q) = %q, %v; want %q", input, path, err, want)
		}
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("normalization created the vault: %v", err)
		}
	}
}

func TestNormalizeVaultPathRejectsOtherUsersAndFiles(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{" ", "~someone/vault", file, filepath.Join(file, "nested")} {
		if _, err := config.NormalizeVaultPath(input); err == nil {
			t.Fatalf("invalid vault path accepted: %q", input)
		}
	}
	content, err := os.ReadFile(file)
	if err != nil || string(content) != "keep" {
		t.Fatalf("path validation altered the existing file: %q, %v", content, err)
	}
}
