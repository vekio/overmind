package main

import (
	"bytes"
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc"
	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/edit"
)

func TestSectionBodyReplacementUsesASTSpanAndDocumentLineEnding(t *testing.T) {
	source := []byte("= Note\r\n\r\n== Summary\r\n\r\nOld text.\r\n\r\n=== Child\r\n\r\nNested.\r\n\r\n== Notes\r\n\r\nKeep.\r\n")
	section := findExampleSection(t, source, "Summary")
	change, err := edit.ReplaceSectionContent(source, section, []byte("New text.\n\nSecond paragraph."))
	if err != nil {
		t.Fatalf("ReplaceSectionContent() error = %v", err)
	}
	updated, err := edit.Apply(source, []edit.TextEdit{change})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	want := "== Summary\r\n\r\nNew text.\r\n\r\nSecond paragraph.\r\n\r\n== Notes"
	if !bytes.Contains(updated, []byte(want)) || bytes.Contains(updated, []byte("Old text")) || bytes.Contains(updated, []byte("=== Child")) || !bytes.Contains(updated, []byte("Keep.")) {
		t.Fatalf("updated source =\n%s", updated)
	}
	if parsed := asciidoc.Parse(updated); parsed.HasErrors() {
		t.Fatalf("reparsed diagnostics = %+v", parsed.Diagnostics)
	}
}

func TestSectionBodyTerminatesHeadingAtEOF(t *testing.T) {
	source := []byte("== Summary")
	section := findExampleSection(t, source, "Summary")
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

func findExampleSection(t *testing.T, source []byte, title string) *ast.Section {
	t.Helper()
	document := asciidoc.Parse(source).Document
	var result *ast.Section
	ast.Walk(document, func(node ast.Node) bool {
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
