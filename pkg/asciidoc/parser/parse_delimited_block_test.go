package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
)

func TestParseDelimitedBlockTreatsDifferentFenceAsContentAndReportsMissingClose(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "unclosed_block.adoc"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	result := parse(source)
	if !result.HasErrors() || len(result.Diagnostics) != 1 {
		t.Fatalf("Diagnostics = %+v, want one error", result.Diagnostics)
	}
	item := result.Diagnostics[0]
	if item.Severity != diagnostic.SeverityError || item.Message != `unclosed listing block; expected "----"` || item.Source.Start.Line != 4 {
		t.Fatalf("diagnostic = %+v", item)
	}
	section := requireBlock[*ast.Section](t, result.Document.Blocks[0])
	block := requireBlock[*ast.DelimitedBlock](t, section.Blocks[0])
	if block.Closed || block.ClosingSource != (ast.Span{}) || block.Content != "puts \"hello\"\n-----\n" || block.Source.End.Offset != len(source) {
		t.Fatalf("partial delimited block = %+v", block)
	}
}

func TestParseDelimitedBlockKindsAndContentModels(t *testing.T) {
	for _, test := range []struct {
		marker string
		kind   ast.DelimitedBlockKind
		model  ast.ContentModel
	}{
		{marker: "--", kind: ast.DelimitedBlockOpen, model: ast.ContentModelCompound},
		{marker: "----", kind: ast.DelimitedBlockListing, model: ast.ContentModelVerbatim},
		{marker: "....", kind: ast.DelimitedBlockLiteral, model: ast.ContentModelVerbatim},
		{marker: "====", kind: ast.DelimitedBlockExample, model: ast.ContentModelCompound},
		{marker: "****", kind: ast.DelimitedBlockSidebar, model: ast.ContentModelCompound},
		{marker: "____", kind: ast.DelimitedBlockQuote, model: ast.ContentModelCompound},
		{marker: "++++", kind: ast.DelimitedBlockPassthrough, model: ast.ContentModelRaw},
	} {
		t.Run(test.kind.String(), func(t *testing.T) {
			result := parse([]byte(test.marker + "\ncontent\n" + test.marker))
			if len(result.Diagnostics) != 0 || len(result.Document.Blocks) != 1 {
				t.Fatalf("result = %+v", result)
			}
			block := requireBlock[*ast.DelimitedBlock](t, result.Document.Blocks[0])
			if block.Kind != test.kind || block.ContentModel != test.model || block.Content != "content\n" || !block.Closed {
				t.Fatalf("block = %+v", block)
			}
		})
	}
}

func TestParseCommentBlockProducesNoNodeAndDoesNotParseItsContent(t *testing.T) {
	result := parse([]byte("////\n=== hidden\n* hidden list\n////\nvisible"))
	if len(result.Diagnostics) != 0 || len(result.Document.Blocks) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if paragraph := requireBlock[*ast.Paragraph](t, result.Document.Blocks[0]); paragraph.Text != "visible" {
		t.Fatalf("paragraph = %+v", paragraph)
	}
}
