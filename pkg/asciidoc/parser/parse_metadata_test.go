package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

func TestParseDelimitedBlocksWithMetadataAndExactContent(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "delimited_blocks.adoc"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	result := parse(source)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v, want none", result.Diagnostics)
	}
	if len(result.Document.Blocks) != 2 {
		t.Fatalf("root blocks = %d, want title section content plus table", len(result.Document.Blocks))
	}

	listing := requireBlock[*ast.DelimitedBlock](t, result.Document.Blocks[0])
	if listing.Kind != ast.DelimitedBlockListing || listing.ContentModel != ast.ContentModelVerbatim || listing.Marker != "-----" || !listing.Closed {
		t.Fatalf("listing identity = %+v", listing)
	}
	wantContent := "package main\n== This is listing content\n// This is also listing content\nimage::diagram.svg[]\n"
	if listing.Content != wantContent {
		t.Fatalf("listing content = %q, want %q", listing.Content, wantContent)
	}
	if listing.Source.Start.Line != 3 || listing.OpeningSource.Start.Line != 7 || listing.ContentSource.Start.Line != 8 || listing.ContentSource.End.Line != 12 || listing.ClosingSource.Start.Line != 12 {
		t.Fatalf("listing spans = source %s opening %s content %s closing %s", listing.Source, listing.OpeningSource, listing.ContentSource, listing.ClosingSource)
	}
	metadata := listing.Metadata
	if metadata.Title == nil || metadata.Title.Text != "Example source" || metadata.Title.TitleSource.Start.Column != 2 {
		t.Fatalf("block title = %+v", metadata.Title)
	}
	if metadata.Anchor == nil || metadata.Anchor.ID != "hello" || metadata.Anchor.RefText != "Hello example" || metadata.Anchor.IDSource.Start.Column != 3 || metadata.Anchor.RefTextSource.Start.Column != 9 {
		t.Fatalf("anchor = %+v", metadata.Anchor)
	}
	if len(metadata.AttributeLists) != 2 || len(metadata.AttributeLists[0].Entries) != 2 || metadata.AttributeLists[0].Entries[1].Value != "go" || metadata.AttributeLists[0].Entries[1].Source.Start.Column != 9 {
		t.Fatalf("attribute lists = %+v", metadata.AttributeLists)
	}

	table := requireBlock[*ast.Table](t, result.Document.Blocks[1])
	if table.Format != ast.TableFormatPSV || table.Content != "|Name |Value\n" || len(table.Columns) != 2 || len(table.Rows) != 1 || !table.Closed {
		t.Fatalf("table = %+v", table)
	}
}

func TestParseMetadataDiagnosticsAndClosestValue(t *testing.T) {
	result := parse([]byte(".First\n.Second\n[[one]]\n[[two]]\n----\ncontent\n----"))
	if result.HasErrors() || len(result.Diagnostics) != 2 {
		t.Fatalf("Diagnostics = %+v, want duplicate warnings", result.Diagnostics)
	}
	block := requireBlock[*ast.DelimitedBlock](t, result.Document.Blocks[0])
	if block.Metadata.Title == nil || block.Metadata.Title.Text != "Second" || block.Metadata.Anchor == nil || block.Metadata.Anchor.ID != "two" {
		t.Fatalf("metadata = %+v", block.Metadata)
	}

	orphan := parse([]byte(".Orphan\n[source,go]\n\nplain"))
	if orphan.HasErrors() || len(orphan.Diagnostics) != 2 || len(orphan.Document.Blocks) != 1 {
		t.Fatalf("orphan result = %+v", orphan)
	}
}
