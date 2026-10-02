package parser_test

import (
	"testing"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
)

func TestParseNestedListAndContinuationBlocks(t *testing.T) {
	source := []byte(".Steps\n* first *item*\n** nested one\n** nested two\n* second\n+\nA continued _paragraph_.\n+\n[source,go]\n----\ncode\n----\n")
	result := parse(source)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	list := requireBlock[*ast.List](t, result.Document.Blocks[0])
	if list.Kind != ast.ListUnordered || list.Level != 1 || len(list.Items) != 2 || list.Metadata.Title == nil || list.Metadata.Title.Text != "Steps" {
		t.Fatalf("list = %+v", list)
	}
	first := list.Items[0]
	if first.Principal != "first *item*" || len(first.Inlines) != 2 || len(first.Blocks) != 1 {
		t.Fatalf("first item = %+v", first)
	}
	nested := requireBlock[*ast.List](t, first.Blocks[0])
	if nested.Level != 2 || len(nested.Items) != 2 {
		t.Fatalf("nested list = %+v", nested)
	}
	second := list.Items[1]
	if len(second.Blocks) != 2 {
		t.Fatalf("second blocks = %+v", second.Blocks)
	}
	paragraph := requireBlock[*ast.Paragraph](t, second.Blocks[0])
	if paragraph.Text != "A continued _paragraph_." || len(paragraph.Inlines) != 3 {
		t.Fatalf("continued paragraph = %+v", paragraph)
	}
	listing := requireBlock[*ast.DelimitedBlock](t, second.Blocks[1])
	if listing.Kind != ast.DelimitedBlockListing || listing.Content != "code\n" {
		t.Fatalf("continued listing = %+v", listing)
	}
}

func TestParseOrderedListAndLevelJumpDiagnostic(t *testing.T) {
	result := parse([]byte(". first\n+\nTIP: attached\n. second\n... jumped"))
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Message != "list level jumped from 1 to 3" {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	list := requireBlock[*ast.List](t, result.Document.Blocks[0])
	if list.Kind != ast.ListOrdered || len(list.Items) != 2 {
		t.Fatalf("list = %+v", list)
	}
	requireBlock[*ast.Admonition](t, list.Items[0].Blocks[0])
	nested := requireBlock[*ast.List](t, list.Items[1].Blocks[0])
	if nested.Level != 3 {
		t.Fatalf("nested level = %d", nested.Level)
	}
}

func TestParseOrphanContinuationWarnsAndRecovers(t *testing.T) {
	result := parse([]byte("+\nplain"))
	if len(result.Diagnostics) != 1 || len(result.Document.Blocks) != 1 {
		t.Fatalf("result = %+v", result)
	}
}
