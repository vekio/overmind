package parser_test

import (
	"testing"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
)

func TestParseParagraphInlineSpansFollowUnicodeAndCRLFSource(t *testing.T) {
	source := []byte("Café *bold\r\nand _deep_* after")
	result := parse(source)
	paragraph := requireBlock[*ast.Paragraph](t, result.Document.Blocks[0])
	if paragraph.Text != "Café *bold\nand _deep_* after" {
		t.Fatalf("paragraph text = %q", paragraph.Text)
	}
	strong := requireInline[*ast.Strong](t, paragraph.Inlines[1])
	if strong.Source.Start != (ast.Position{Offset: 6, Line: 1, Column: 6}) || strong.Source.End.Line != 2 {
		t.Fatalf("strong source = %s", strong.Source)
	}
	if string(source[strong.Source.Start.Offset:strong.Source.End.Offset]) != "*bold\r\nand _deep_*" {
		t.Fatalf("strong source bytes = %q", source[strong.Source.Start.Offset:strong.Source.End.Offset])
	}
	if inlineText(t, strong.Children) != "bold\nand deep" {
		t.Fatalf("strong children = %#v", strong.Children)
	}
	emphasis := requireInline[*ast.Emphasis](t, strong.Children[1])
	if string(source[emphasis.ContentSource.Start.Offset:emphasis.ContentSource.End.Offset]) != "deep" {
		t.Fatalf("emphasis content source = %s", emphasis.ContentSource)
	}
}
