package parser_test

import (
	"testing"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
)

func TestParseParagraphEscapedFormattingMarksRemainText(t *testing.T) {
	source := `\*literal\* \_value\_ \` + "`code\\` \\**characters**"
	result := parse([]byte(source))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v, want none", result.Diagnostics)
	}
	paragraph := requireBlock[*ast.Paragraph](t, result.Document.Blocks[0])
	if len(paragraph.Inlines) != 1 {
		t.Fatalf("inlines = %#v, want one text node", paragraph.Inlines)
	}
	text := requireInline[*ast.Text](t, paragraph.Inlines[0])
	if text.Value != "*literal* _value_ `code` **characters**" || text.Source != paragraph.Source {
		t.Fatalf("text = %+v", text)
	}
}
