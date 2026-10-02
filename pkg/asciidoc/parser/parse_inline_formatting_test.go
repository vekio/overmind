package parser_test

import (
	"testing"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
)

func TestParseParagraphConstrainedInlineFormatting(t *testing.T) {
	result := parse([]byte("Use *strong*, _emphasis_, and `monospace`."))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v, want none", result.Diagnostics)
	}
	paragraph := requireBlock[*ast.Paragraph](t, result.Document.Blocks[0])
	if len(paragraph.Inlines) != 7 {
		t.Fatalf("inlines = %#v, want seven nodes", paragraph.Inlines)
	}
	if text := requireInline[*ast.Text](t, paragraph.Inlines[0]); text.Value != "Use " {
		t.Fatalf("first text = %+v", text)
	}
	strong := requireInline[*ast.Strong](t, paragraph.Inlines[1])
	if inlineText(t, strong.Children) != "strong" || sourceText(t, paragraph.Text, strong.Source) != "*strong*" || sourceText(t, paragraph.Text, strong.ContentSource) != "strong" {
		t.Fatalf("strong = %+v", strong)
	}
	emphasis := requireInline[*ast.Emphasis](t, paragraph.Inlines[3])
	if inlineText(t, emphasis.Children) != "emphasis" {
		t.Fatalf("emphasis = %+v", emphasis)
	}
	monospace := requireInline[*ast.Monospace](t, paragraph.Inlines[5])
	if inlineText(t, monospace.Children) != "monospace" {
		t.Fatalf("monospace = %+v", monospace)
	}
}

func TestParseParagraphNestedInlineFormatting(t *testing.T) {
	result := parse([]byte("Use `*_all_*` together."))
	paragraph := requireBlock[*ast.Paragraph](t, result.Document.Blocks[0])
	monospace := requireInline[*ast.Monospace](t, paragraph.Inlines[1])
	strong := requireInline[*ast.Strong](t, monospace.Children[0])
	emphasis := requireInline[*ast.Emphasis](t, strong.Children[0])
	if inlineText(t, emphasis.Children) != "all" {
		t.Fatalf("nested inlines = %#v", paragraph.Inlines)
	}
}

func TestParseParagraphUnclosedFormattingPairsRemainText(t *testing.T) {
	for _, test := range []struct {
		name   string
		source string
	}{
		{name: "unclosed", source: "an *unclosed marker"},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := parse([]byte(test.source))
			if len(result.Diagnostics) != 0 {
				t.Fatalf("Diagnostics = %+v, want none", result.Diagnostics)
			}
			paragraph := requireBlock[*ast.Paragraph](t, result.Document.Blocks[0])
			if len(paragraph.Inlines) != 1 {
				t.Fatalf("inlines = %#v, want one text node", paragraph.Inlines)
			}
			text := requireInline[*ast.Text](t, paragraph.Inlines[0])
			if text.Value != test.source || text.Source != paragraph.Source {
				t.Fatalf("text = %+v, want value %q and source %s", text, test.source, paragraph.Source)
			}
		})
	}
}

func TestParseParagraphUnconstrainedInlineFormatting(t *testing.T) {
	result := parse([]byte("**C**reate fan__tastic__ and re``format``ted"))
	paragraph := requireBlock[*ast.Paragraph](t, result.Document.Blocks[0])
	if len(paragraph.Inlines) != 6 {
		t.Fatalf("inlines = %#v, want six nodes", paragraph.Inlines)
	}
	if inlineText(t, requireInline[*ast.Strong](t, paragraph.Inlines[0]).Children) != "C" {
		t.Fatalf("strong = %#v", paragraph.Inlines[0])
	}
	if inlineText(t, requireInline[*ast.Emphasis](t, paragraph.Inlines[2]).Children) != "tastic" {
		t.Fatalf("emphasis = %#v", paragraph.Inlines[2])
	}
	if inlineText(t, requireInline[*ast.Monospace](t, paragraph.Inlines[4]).Children) != "format" {
		t.Fatalf("monospace = %#v", paragraph.Inlines[4])
	}
}

func TestParseParagraphAppliesConstrainedBoundaries(t *testing.T) {
	for _, source := range []string{
		"word*not strong*",
		":*not strong*",
		"*not strong*word",
	} {
		result := parse([]byte(source))
		paragraph := requireBlock[*ast.Paragraph](t, result.Document.Blocks[0])
		if len(paragraph.Inlines) != 1 || requireInline[*ast.Text](t, paragraph.Inlines[0]).Value != source {
			t.Errorf("%q inlines = %#v, want literal text", source, paragraph.Inlines)
		}
	}

	result := parse([]byte("(*strong*) and *9*-to-*5*"))
	paragraph := requireBlock[*ast.Paragraph](t, result.Document.Blocks[0])
	if len(paragraph.Inlines) != 6 {
		t.Fatalf("punctuation-bound inlines = %#v", paragraph.Inlines)
	}
	for _, index := range []int{1, 3, 5} {
		requireInline[*ast.Strong](t, paragraph.Inlines[index])
	}
}
