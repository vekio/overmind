package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParserCommandPrintsResultAsJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.adoc")
	source := "= Example\n\nIntroduction.\n\n== Installation\n\nRun this.\n\n'''\n"
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var output bytes.Buffer
	command := Command()
	command.Writer = &output
	command.ErrWriter = &output
	if err := command.Run(context.Background(), []string{"asciidoc", "parser", path}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var result testJSONResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output:\n%s", err, output.String())
	}
	if result.Document.Kind != "document" || result.Document.Title == nil || result.Document.Title.Text != "Example" {
		t.Fatalf("document = %+v", result.Document)
	}
	if result.Document.Source.End.Offset != len(source) {
		t.Fatalf("document end offset = %d, want %d", result.Document.Source.End.Offset, len(source))
	}
	if len(result.Document.Blocks) != 2 || result.Document.Blocks[0].Kind != "paragraph" {
		t.Fatalf("root blocks = %+v", result.Document.Blocks)
	}
	section := result.Document.Blocks[1]
	if section.Kind != "section" || section.Level == nil || *section.Level != 1 || section.Title == nil || *section.Title != "Installation" {
		t.Fatalf("section = %+v", section)
	}
	if section.Blocks == nil || len(*section.Blocks) != 2 || (*section.Blocks)[0].Kind != "paragraph" || (*section.Blocks)[1].Kind != "thematic_break" {
		t.Fatalf("section blocks = %+v", section.Blocks)
	}
	if result.Diagnostics == nil || len(result.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want empty JSON array", result.Diagnostics)
	}
	if !strings.HasPrefix(output.String(), "{\n  \"document\"") {
		t.Fatalf("output is not indented:\n%s", output.String())
	}
}

func TestParserCommandPrintsWarningsAndSucceeds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "warning.adoc")
	if err := os.WriteFile(path, []byte("=== Too deep\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var output bytes.Buffer
	command := Command()
	command.Writer = &output
	if err := command.Run(context.Background(), []string{"asciidoc", "parser", path}); err != nil {
		t.Fatalf("Run() warning error = %v", err)
	}
	var result testJSONResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Severity != "warning" {
		t.Fatalf("diagnostics = %+v, want one warning", result.Diagnostics)
	}
}

func TestParserCommandPrintsPartialResultBeforeReturningError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.adoc")
	if err := os.WriteFile(path, []byte("body\n\n= Late title\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var output bytes.Buffer
	command := Command()
	command.Writer = &output
	err := command.Run(context.Background(), []string{"asciidoc", "parser", path})
	if err == nil || !strings.Contains(err.Error(), "completed with errors") {
		t.Fatalf("Run() error = %v, want parser error", err)
	}
	var result testJSONResult
	if jsonErr := json.Unmarshal(output.Bytes(), &result); jsonErr != nil {
		t.Fatalf("JSON error = %v; output:\n%s", jsonErr, output.String())
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Severity != "error" || len(result.Document.Blocks) != 1 {
		t.Fatalf("partial result = %+v", result)
	}
}

func TestParserCommandValidatesPathAndReportsWriteError(t *testing.T) {
	command := Command()
	if err := command.Run(context.Background(), []string{"asciidoc", "parser", "document.txt"}); err == nil || !strings.Contains(err.Error(), ".adoc extension") {
		t.Fatalf("extension error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "document.adoc")
	if err := os.WriteFile(path, []byte("plain"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	command = Command()
	command.Writer = errorWriter{err: errors.New("broken output")}
	if err := command.Run(context.Background(), []string{"asciidoc", "parser", path}); err == nil || !strings.Contains(err.Error(), "broken output") {
		t.Fatalf("write error = %v", err)
	}
}

type errorWriter struct{ err error }

func (w errorWriter) Write([]byte) (int, error) { return 0, w.err }

type testJSONResult struct {
	Document    testJSONDocument     `json:"document"`
	Diagnostics []testJSONDiagnostic `json:"diagnostics"`
}

type testJSONDocument struct {
	Kind   string                 `json:"kind"`
	Source testJSONSpan           `json:"source"`
	Title  *testJSONDocumentTitle `json:"title"`
	Blocks []testJSONNode         `json:"blocks"`
}

type testJSONDocumentTitle struct {
	Text string `json:"text"`
}

type testJSONNode struct {
	Kind   string          `json:"kind"`
	Level  *int            `json:"level"`
	Title  *string         `json:"title"`
	Blocks *[]testJSONNode `json:"blocks"`
}

type testJSONDiagnostic struct {
	Severity string `json:"severity"`
}

type testJSONSpan struct {
	End testJSONPosition `json:"end"`
}

type testJSONPosition struct {
	Offset int `json:"offset"`
}
