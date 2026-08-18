package local

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/config"
)

func TestAsciiDocCommandDoesNotLoadApplicationConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.adoc")
	if err := os.WriteFile(path, []byte("plain text\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	built := false
	command := newTestCommand(t, func(config.CLIConfig) (Runtime, error) {
		built = true
		return runtimeStub{application: &app.Application{}}, nil
	})
	command.Writer = &bytes.Buffer{}
	if err := command.Run(context.Background(), []string{"overmind", "asciidoc", "lexer", path}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if built {
		t.Fatal("application was built for standalone asciidoc command")
	}
}

func TestConfigInitDoesNotBuildApplication(t *testing.T) {
	configHome := filepath.Join(t.TempDir(), "config")
	dataHome := filepath.Join(t.TempDir(), "data")
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_DATA_HOME", dataHome)

	built := false
	command := newTestCommand(t, func(config.CLIConfig) (Runtime, error) {
		built = true
		return runtimeStub{application: &app.Application{}}, nil
	})
	var output bytes.Buffer
	command.Writer = &output

	if err := command.Run(context.Background(), []string{"overmind", "config", "init"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	wantPath := filepath.Join(configHome, "overmind", "config.yml")
	if output.String() != wantPath+"\n" {
		t.Fatalf("output = %q, want %q", output.String(), wantPath+"\n")
	}
	if built {
		t.Fatal("application was built for config command")
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
}
