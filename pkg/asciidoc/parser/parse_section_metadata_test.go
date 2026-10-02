package parser_test

import (
	"testing"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
)

func TestParseAttachesMetadataToSection(t *testing.T) {
	result := parse([]byte("[[summary,Generated summary]]\n.Summary block\n== Summary\n\nPending."))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	section := requireBlock[*ast.Section](t, result.Document.Blocks[0])
	if section.Metadata.Anchor == nil || section.Metadata.Anchor.ID != "summary" || section.Metadata.Anchor.RefText != "Generated summary" {
		t.Fatalf("section metadata = %+v", section.Metadata)
	}
	if section.Metadata.Title == nil || section.Metadata.Title.Text != "Summary block" {
		t.Fatalf("section title metadata = %+v", section.Metadata.Title)
	}
	if section.Source.Start.Line != 1 || section.HeadingSource.Start.Line != 3 || section.HeadingSource.End.Column != 11 || section.TitleSource.Start.Line != 3 {
		t.Fatalf("section source = %s, heading source = %s, title source = %s", section.Source, section.HeadingSource, section.TitleSource)
	}
	if section.ContentSource.Start.Line != 4 || section.ContentSource.End != section.Source.End {
		t.Fatalf("section content source = %s, section source = %s", section.ContentSource, section.Source)
	}
}
