package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAsciiDocCommandDoesNotLoadApplicationConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.adoc")
	if err := os.WriteFile(path, []byte("plain text\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	configHome := filepath.Join(t.TempDir(), "config")
	dataHome := filepath.Join(t.TempDir(), "data")
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_DATA_HOME", dataHome)
	command := newTestCommand(t)
	command.Writer = &bytes.Buffer{}
	if err := command.Run(context.Background(), []string{"overmind", "asciidoc", "lexer", path}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(configHome, "overmind", "config.yml")); !os.IsNotExist(err) {
		t.Fatalf("configuration was created for asciidoc command: %v", err)
	}
}

func TestConfigInitDoesNotBuildApplication(t *testing.T) {
	configHome := filepath.Join(t.TempDir(), "config")
	dataHome := filepath.Join(t.TempDir(), "data")
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_DATA_HOME", dataHome)

	command := newTestCommand(t)
	var output bytes.Buffer
	command.Writer = &output

	if err := command.Run(context.Background(), []string{"overmind", "config", "init"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	wantPath := filepath.Join(configHome, "overmind", "config.yml")
	if output.String() != wantPath+"\n" {
		t.Fatalf("output = %q, want %q", output.String(), wantPath+"\n")
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dataHome, "overmind")); !os.IsNotExist(err) {
		t.Fatalf("data directory was created for config command: %v", err)
	}
}
