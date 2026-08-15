package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLexerCommandPrintsTokens(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.adoc")
	if err := os.WriteFile(path, []byte("== Title\n\n* item\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var output bytes.Buffer
	command := Command()
	command.Writer = &output
	command.ErrWriter = &output
	if err := command.Run(context.Background(), []string{"asciidoc", "lexer", path}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := strings.Join([]string{
		`HEADING("== Title") [1:1@0, 1:9@8)`,
		`BLANK("") [2:1@9, 2:1@9)`,
		`LIST_ITEM("* item") [3:1@10, 3:7@16)`,
		"",
	}, "\n")
	if output.String() != want {
		t.Fatalf("command output = %q, want %q", output.String(), want)
	}
}

func TestLexerCommandRequiresPath(t *testing.T) {
	err := Command().Run(context.Background(), []string{"asciidoc", "lexer"})
	if err == nil || !strings.Contains(err.Error(), "path") {
		t.Fatalf("Run() error = %v, want missing path error", err)
	}
}

func TestLexerCommandRequiresAdocExtension(t *testing.T) {
	err := Command().Run(context.Background(), []string{"asciidoc", "lexer", "document.txt"})
	if err == nil || !strings.Contains(err.Error(), ".adoc extension") {
		t.Fatalf("Run() error = %v, want .adoc extension error", err)
	}
}

func TestLexerCommandReportsOpenError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.adoc")
	err := Command().Run(context.Background(), []string{"asciidoc", "lexer", path})
	if err == nil || !strings.Contains(err.Error(), "open AsciiDoc file") {
		t.Fatalf("Run() error = %v, want open error", err)
	}
}
