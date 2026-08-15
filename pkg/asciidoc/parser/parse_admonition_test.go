package parser_test

import (
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

func TestParseAdmonitionParagraphWithInlineAndContinuationText(t *testing.T) {
	result := parse([]byte(".Remember\nTIP: Use *this* approach.\nIt also spans lines."))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	admonition := requireBlock[*ast.Admonition](t, result.Document.Blocks[0])
	if admonition.Kind != ast.AdmonitionTip || admonition.Label != "TIP" || admonition.Text != "Use *this* approach.\nIt also spans lines." {
		t.Fatalf("admonition = %+v", admonition)
	}
	if admonition.Metadata.Title == nil || admonition.Metadata.Title.Text != "Remember" || inlineText(t, admonition.Inlines) != "Use this approach.\nIt also spans lines." {
		t.Fatalf("admonition metadata/inlines = %+v", admonition)
	}
}
