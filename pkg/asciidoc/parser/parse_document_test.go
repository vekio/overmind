package parser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
)

func TestParseBuildsDocumentParagraphsSectionsAndBreaks(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "basic.adoc"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	result := parse(source)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v, want none", result.Diagnostics)
	}
	document := result.Document
	if document == nil {
		t.Fatal("Document is nil")
	}
	if document.Source.Start != (ast.Position{Line: 1, Column: 1}) || document.Source.End.Offset != len(source) {
		t.Fatalf("Document source = %s, want complete %d-byte input", document.Source, len(source))
	}
	if document.Title == nil || document.Title.Text != "Documento de prueba" {
		t.Fatalf("Document title = %+v, want Documento de prueba", document.Title)
	}
	if document.Title.TitleSource.Start != (ast.Position{Offset: 2, Line: 1, Column: 3}) {
		t.Fatalf("Document title source starts at %s, want 1:3@2", document.Title.TitleSource.Start)
	}

	if len(document.Blocks) != 3 {
		t.Fatalf("root block count = %d, want 3", len(document.Blocks))
	}
	intro := requireBlock[*ast.Paragraph](t, document.Blocks[0])
	if intro.Text != "Introducción en dos líneas,\ncon texto Unicode: café." {
		t.Fatalf("intro text = %q", intro.Text)
	}
	first := requireBlock[*ast.Section](t, document.Blocks[1])
	second := requireBlock[*ast.Section](t, document.Blocks[2])
	if first.Level != 1 || first.Title != "Primera sección" || second.Level != 1 || second.Title != "Segunda sección" {
		t.Fatalf("sections = (%d %q), (%d %q)", first.Level, first.Title, second.Level, second.Title)
	}
	if first.Source.End != second.Source.Start {
		t.Fatalf("first section ends at %s, second starts at %s", first.Source.End, second.Source.Start)
	}
	if len(first.Blocks) != 2 {
		t.Fatalf("first section block count = %d, want 2", len(first.Blocks))
	}
	paragraph := requireBlock[*ast.Paragraph](t, first.Blocks[0])
	if paragraph.Text != "Contenido principal." {
		t.Fatalf("section paragraph = %q", paragraph.Text)
	}
	child := requireBlock[*ast.Section](t, first.Blocks[1])
	if child.Level != 2 || child.Title != "Sección hija" || len(child.Blocks) != 1 {
		t.Fatalf("child section = %+v", child)
	}
	requireBlock[*ast.ThematicBreak](t, child.Blocks[0])
	if len(second.Blocks) != 1 {
		t.Fatalf("second section block count = %d, want 1", len(second.Blocks))
	}
	requireBlock[*ast.PageBreak](t, second.Blocks[0])
}

func TestParseCommentsSeparateParagraphsWithoutCreatingNodes(t *testing.T) {
	result := parse([]byte("first\n// separator\nsecond"))
	if len(result.Diagnostics) != 0 || len(result.Document.Blocks) != 2 {
		t.Fatalf("result = %+v, want two paragraphs and no diagnostics", result)
	}
	if requireBlockText(t, result.Document.Blocks[0]) != "first" || requireBlockText(t, result.Document.Blocks[1]) != "second" {
		t.Fatalf("paragraphs = %+v", result.Document.Blocks)
	}
}

func TestParseWarnsAndRecoversFromUnsupportedLine(t *testing.T) {
	result := parse([]byte("+\nplain"))
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Severity != diagnostic.SeverityWarning {
		t.Fatalf("Diagnostics = %+v, want one warning", result.Diagnostics)
	}
	if !strings.Contains(result.Diagnostics[0].Message, "CONTINUATION") {
		t.Fatalf("diagnostic message = %q, want CONTINUATION", result.Diagnostics[0].Message)
	}
	if len(result.Document.Blocks) != 1 || requireBlockText(t, result.Document.Blocks[0]) != "plain" {
		t.Fatalf("partial AST blocks = %+v", result.Document.Blocks)
	}
}
