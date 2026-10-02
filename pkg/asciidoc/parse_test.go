package asciidoc_test

import (
	"strings"
	"testing"

	"github.com/vekio/overmind/pkg/asciidoc"
	"github.com/vekio/overmind/pkg/asciidoc/ast"
)

func TestParseFacade(t *testing.T) {
	result := asciidoc.Parse([]byte("= Title\n\nbody"))
	if result.HasErrors() || result.Document.Title == nil || result.Document.Title.Text != "Title" {
		t.Fatalf("Parse() result = %+v", result)
	}
	if len(result.Document.Blocks) != 1 {
		t.Fatalf("Parse() blocks = %d, want 1", len(result.Document.Blocks))
	}
	if _, ok := result.Document.Blocks[0].(*ast.Paragraph); !ok {
		t.Fatalf("Parse() block type = %T, want *ast.Paragraph", result.Document.Blocks[0])
	}
}

func TestParseReaderFacade(t *testing.T) {
	result := asciidoc.ParseReader(strings.NewReader("plain"))
	if result.HasErrors() || result.Document.Source.End.Offset != len("plain") {
		t.Fatalf("ParseReader() result = %+v", result)
	}
}
