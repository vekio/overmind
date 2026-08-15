package edit_test

import (
	"bytes"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc"
	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/edit"
)

func TestReplaceSectionContentReplacesCompleteBodyAndNestedSections(t *testing.T) {
	source := []byte("= Note\n\n== Summary\n\nOld text.\n\n=== Child\n\nNested.\n\n== Notes\n\nKeep.\n")
	section := findSection(t, source, "Summary")
	change, err := edit.ReplaceSectionContent(source, section, []byte("New text.\n\nSecond paragraph."))
	if err != nil {
		t.Fatalf("ReplaceSectionContent() error = %v", err)
	}
	updated, err := edit.Apply(source, []edit.TextEdit{change})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	want := "== Summary\n\nNew text.\n\nSecond paragraph.\n\n== Notes"
	if !bytes.Contains(updated, []byte(want)) || bytes.Contains(updated, []byte("Old text")) || bytes.Contains(updated, []byte("=== Child")) || !bytes.Contains(updated, []byte("Keep.")) {
		t.Fatalf("updated source =\n%s", updated)
	}
	if parsed := asciidoc.Parse(updated); parsed.HasErrors() {
		t.Fatalf("reparsed diagnostics = %+v", parsed.Diagnostics)
	}
}

func TestReplaceSectionContentUsesHeadingLineEndingInMixedSource(t *testing.T) {
	source := []byte("= Mixed\n\n== Earlier\n\nText.\n\n== Summary\r\n\r\nOld.\r\n")
	section := findSection(t, source, "Summary")
	change, err := edit.ReplaceSectionContent(source, section, []byte("Línea uno.\nLínea dos."))
	if err != nil {
		t.Fatalf("ReplaceSectionContent() error = %v", err)
	}
	if strings.Contains(strings.ReplaceAll(string(change.Replacement), "\r\n", ""), "\n") {
		t.Fatalf("replacement contains bare LF: %q", change.Replacement)
	}
	if !bytes.Contains(change.Replacement, []byte("Línea uno.\r\nLínea dos.")) {
		t.Fatalf("replacement = %q", change.Replacement)
	}
}

func TestReplaceSectionContentTerminatesHeadingAtEOF(t *testing.T) {
	source := []byte("== Summary")
	section := findSection(t, source, "Summary")
	change, err := edit.ReplaceSectionContent(source, section, []byte("Generated."))
	if err != nil {
		t.Fatalf("ReplaceSectionContent() error = %v", err)
	}
	updated, err := edit.Apply(source, []edit.TextEdit{change})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if string(updated) != "== Summary\n\nGenerated.\n\n" {
		t.Fatalf("updated = %q", updated)
	}
}

func TestReplaceSectionContentSupportsEmptyBody(t *testing.T) {
	source := []byte("== Summary\n\nOld.\n\n== Next\n")
	section := findSection(t, source, "Summary")
	change, err := edit.ReplaceSectionContent(source, section, nil)
	if err != nil {
		t.Fatalf("ReplaceSectionContent() error = %v", err)
	}
	updated, err := edit.Apply(source, []edit.TextEdit{change})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if string(updated) != "== Summary\n\n== Next\n" {
		t.Fatalf("updated = %q", updated)
	}
}

func TestReplaceSectionContentRejectsNilAndInvalidSpans(t *testing.T) {
	if _, err := edit.ReplaceSectionContent(nil, nil, nil); err == nil {
		t.Fatal("ReplaceSectionContent(nil) returned no error")
	}
	section := &ast.Section{
		HeadingSource: ast.Span{Start: ast.Position{Offset: 0}, End: ast.Position{Offset: 2}},
		ContentSource: ast.Span{Start: ast.Position{Offset: 3}, End: ast.Position{Offset: 8}},
	}
	if _, err := edit.ReplaceSectionContent([]byte("invalid"), section, nil); err == nil {
		t.Fatal("ReplaceSectionContent() returned no error for invalid spans")
	}
}

func findSection(t *testing.T, source []byte, title string) *ast.Section {
	t.Helper()
	parsed := asciidoc.Parse(source)
	if parsed.HasErrors() {
		t.Fatalf("Parse() diagnostics = %+v", parsed.Diagnostics)
	}
	var result *ast.Section
	ast.Walk(parsed.Document, func(node ast.Node) bool {
		section, ok := node.(*ast.Section)
		if ok && section.Title == title {
			result = section
			return false
		}
		return true
	})
	if result == nil {
		t.Fatalf("section %q not found", title)
	}
	return result
}
