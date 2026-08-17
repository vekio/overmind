package cli

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
	command := NewCommand(func(config.Config) (Runtime, error) {
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
