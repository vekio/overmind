package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/diagnostic"
)

func TestParseReportsSectionLevelJumpAndKeepsPartialTree(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "section_jump.adoc"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	result := parse(source)
	if result.HasErrors() {
		t.Fatalf("HasErrors() = true for warning: %+v", result.Diagnostics)
	}
	if len(result.Diagnostics) != 1 {
		t.Fatalf("diagnostic count = %d, want 1", len(result.Diagnostics))
	}
	warning := result.Diagnostics[0]
	if warning.Severity != diagnostic.SeverityWarning || warning.Message != "section level jumped from 0 to 2" {
		t.Fatalf("diagnostic = %+v", warning)
	}
	if warning.Source.Start.Line != 1 || warning.Source.End.Column != 4 {
		t.Fatalf("diagnostic source = %s, want heading marker", warning.Source)
	}

	if len(result.Document.Blocks) != 2 {
		t.Fatalf("root blocks = %d, want 2", len(result.Document.Blocks))
	}
	tooDeep := requireBlock[*ast.Section](t, result.Document.Blocks[0])
	if tooDeep.Level != 2 || len(tooDeep.Blocks) != 1 {
		t.Fatalf("first section = %+v", tooDeep)
	}
	child := requireBlock[*ast.Section](t, tooDeep.Blocks[0])
	if child.Level != 3 {
		t.Fatalf("child level = %d, want 3", child.Level)
	}
	back := requireBlock[*ast.Section](t, result.Document.Blocks[1])
	if back.Level != 1 {
		t.Fatalf("back level = %d, want 1", back.Level)
	}
}

func TestParseSectionContentSourceIncludesBodyAndNestedSections(t *testing.T) {
	source := []byte("== Parent\r\n\r\nBody.\r\n\r\n=== Child\r\n\r\nNested.\r\n\r\n[[next]]\r\n== Next\r\n")
	result := parse(source)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	parent := requireBlock[*ast.Section](t, result.Document.Blocks[0])
	want := "\r\nBody.\r\n\r\n=== Child\r\n\r\nNested.\r\n\r\n"
	if got := string(source[parent.ContentSource.Start.Offset:parent.ContentSource.End.Offset]); got != want {
		t.Fatalf("parent content = %q, want %q", got, want)
	}
	if parent.ContentSource.Start.Line != 2 || parent.ContentSource.End.Line != 9 {
		t.Fatalf("parent content source = %s", parent.ContentSource)
	}
	next := requireBlock[*ast.Section](t, result.Document.Blocks[1])
	if parent.ContentSource.End != next.Source.Start || next.ContentSource.Start.Offset != len(source) {
		t.Fatalf("parent content = %s, next source = %s, next content = %s", parent.ContentSource, next.Source, next.ContentSource)
	}
}
