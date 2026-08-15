package parser_test

import (
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

func TestParseParagraphLinksAndCrossReferences(t *testing.T) {
	result := parse([]byte("Visit https://example.com[*Example*], link:manual.pdf[Manual], <<install,Install>>, and xref:guide.adoc[Guide]."))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	paragraph := requireBlock[*ast.Paragraph](t, result.Document.Blocks[0])
	link := requireInline[*ast.Link](t, paragraph.Inlines[1])
	if link.Target != "https://example.com" || len(link.Children) != 1 {
		t.Fatalf("URL link = %+v", link)
	}
	requireInline[*ast.Strong](t, link.Children[0])
	fileLink := requireInline[*ast.Link](t, paragraph.Inlines[3])
	if fileLink.Target != "manual.pdf" || inlineText(t, fileLink.Children) != "Manual" {
		t.Fatalf("file link = %+v", fileLink)
	}
	short := requireInline[*ast.CrossReference](t, paragraph.Inlines[5])
	if short.Target != "install" || inlineText(t, short.Children) != "Install" {
		t.Fatalf("short xref = %+v", short)
	}
	xref := requireInline[*ast.CrossReference](t, paragraph.Inlines[7])
	if xref.Target != "guide.adoc" || inlineText(t, xref.Children) != "Guide" {
		t.Fatalf("xref = %+v", xref)
	}
}

func TestParseParagraphEmptyLabelsAndEscapedMacros(t *testing.T) {
	result := parse([]byte(`https://example.com[] <<target>> \https://literal.example[Text]`))
	paragraph := requireBlock[*ast.Paragraph](t, result.Document.Blocks[0])
	link := requireInline[*ast.Link](t, paragraph.Inlines[0])
	if link.Target != "https://example.com" || len(link.Children) != 0 {
		t.Fatalf("empty link = %+v", link)
	}
	xref := requireInline[*ast.CrossReference](t, paragraph.Inlines[2])
	if xref.Target != "target" || len(xref.Children) != 0 {
		t.Fatalf("empty xref = %+v", xref)
	}
	if text := requireInline[*ast.Text](t, paragraph.Inlines[3]); text.Value != " https://literal.example[Text]" {
		t.Fatalf("escaped macro text = %+v", text)
	}
}

func TestParseParagraphDoesNotMatchMacroInsideWord(t *testing.T) {
	source := "prefixhttps://example.com[Text]"
	result := parse([]byte(source))
	paragraph := requireBlock[*ast.Paragraph](t, result.Document.Blocks[0])
	if len(paragraph.Inlines) != 1 || requireInline[*ast.Text](t, paragraph.Inlines[0]).Value != source {
		t.Fatalf("inlines = %#v, want literal text", paragraph.Inlines)
	}
}
