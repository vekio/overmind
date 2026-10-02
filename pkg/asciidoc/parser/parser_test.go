package parser_test

import (
	"errors"
	"strings"
	"testing"
	"testing/iotest"

	parserpkg "github.com/vekio/overmind/pkg/asciidoc/parser"
)

func TestParseReaderReturnsPartialDocumentAndReadDiagnostic(t *testing.T) {
	reader := iotest.TimeoutReader(strings.NewReader("partial"))
	result := parseReader(reader)
	if !result.HasErrors() || len(result.Diagnostics) != 1 || !strings.Contains(result.Diagnostics[0].Message, iotest.ErrTimeout.Error()) {
		t.Fatalf("Diagnostics = %+v, want timeout error", result.Diagnostics)
	}
	if len(result.Document.Blocks) != 1 || requireBlockText(t, result.Document.Blocks[0]) != "partial" {
		t.Fatalf("partial document = %+v", result.Document)
	}
}

func TestParserRejectsNilReaderAndReuse(t *testing.T) {
	parser := parserpkg.New(nil)
	first := parser.Parse()
	if !first.HasErrors() || first.Document == nil {
		t.Fatalf("first result = %+v, want error and document", first)
	}
	second := parser.Parse()
	if !second.HasErrors() || second.Diagnostics[0].Message != "parser cannot be reused" {
		t.Fatalf("second result = %+v, want reuse error", second)
	}

	var nilParser *parserpkg.Parser
	if result := nilParser.Parse(); !result.HasErrors() || result.Document == nil {
		t.Fatalf("nil parser result = %+v, want error and document", result)
	}
}

func TestParseReaderPropagatesImmediateReadError(t *testing.T) {
	want := errors.New("broken reader")
	result := parseReader(errorReader{err: want})
	if !result.HasErrors() || !strings.Contains(result.Diagnostics[0].Message, want.Error()) {
		t.Fatalf("Diagnostics = %+v, want broken-reader error", result.Diagnostics)
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }
