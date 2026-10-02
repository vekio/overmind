package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/vekio/overmind/internal/config"
)

func TestNormalizeVaultPathRejectsFilesAndNormalizesInput(t *testing.T) {
	root := t.TempDir()
	path, err := normalizeVaultPath("  " + root + "/  ")
	if err != nil || path != root {
		t.Fatalf("directory path = %q, %v", path, err)
	}
	filePath := filepath.Join(root, "file")
	if err := os.WriteFile(filePath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{" ", filePath, "~someone/vault"} {
		if _, err := normalizeVaultPath(value); err == nil {
			t.Fatalf("invalid vault path %q accepted", value)
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	path, err = normalizeVaultPath("~/overmind-test-vault")
	if err != nil || path != filepath.Join(home, "overmind-test-vault") {
		t.Fatalf("expanded home path = %q, %v", path, err)
	}
}

func TestSetupCreatesConfigOnceWithNormalizedVault(t *testing.T) {
	root := t.TempDir()
	configFile, err := appconfig.NewFile()
	if err != nil {
		t.Fatal(err)
	}
	if err := configFile.SetPath(filepath.Join(root, "config.yml")); err != nil {
		t.Fatal(err)
	}
	called := 0
	prompt := func(context.Context, string) (string, error) {
		called++
		return "  " + filepath.Join(root, "vault") + "/  ", nil
	}
	var output bytes.Buffer
	command := newSetupCommand(configFile, prompt)
	command.Writer = &output
	if err := command.Run(context.Background(), []string{"setup"}); err != nil {
		t.Fatal(err)
	}
	settings, err := configFile.Load()
	if err != nil || settings.Mode != appconfig.ModeLocal || settings.VaultPath != filepath.Join(root, "vault") {
		t.Fatalf("saved settings = %+v, %v", settings, err)
	}
	if output.String() != configFile.Path()+"\n" {
		t.Fatalf("setup output = %q", output.String())
	}
	command = newSetupCommand(configFile, prompt)
	command.Writer = io.Discard
	if err := command.Run(context.Background(), []string{"setup"}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("second setup = %v", err)
	}
	if called != 1 {
		t.Fatalf("prompt called %d times", called)
	}
}
