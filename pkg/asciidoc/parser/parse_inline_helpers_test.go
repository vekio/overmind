package parser_test

import (
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

func requireInline[T ast.Inline](t *testing.T, inline ast.Inline) T {
	t.Helper()
	result, ok := inline.(T)
	if !ok {
		t.Fatalf("inline type = %T, want %T", inline, result)
	}
	return result
}

func inlineText(t *testing.T, inlines []ast.Inline) string {
	t.Helper()
	var result string
	for _, inline := range inlines {
		switch inline := inline.(type) {
		case *ast.Text:
			result += inline.Value
		case *ast.Strong:
			result += inlineText(t, inline.Children)
		case *ast.Emphasis:
			result += inlineText(t, inline.Children)
		case *ast.Monospace:
			result += inlineText(t, inline.Children)
		default:
			t.Fatalf("unsupported inline %T", inline)
		}
	}
	return result
}

func sourceText(t *testing.T, source string, span ast.Span) string {
	t.Helper()
	if span.Start.Line != 1 || span.End.Line != 1 {
		t.Fatalf("sourceText helper only accepts a single line: %s", span)
	}
	return source[span.Start.Offset:span.End.Offset]
}
