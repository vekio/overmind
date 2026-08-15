package edit_test

import (
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/edit"
)

func TestApplyInsertionReplacementAndDeletion(t *testing.T) {
	source := []byte("one two three")
	result, err := edit.Apply(source, []edit.TextEdit{
		{Source: offsetSpan(13, 13), Replacement: []byte("!")},
		{Source: offsetSpan(4, 7), Replacement: []byte("TWO")},
		{Source: offsetSpan(0, 4)},
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if string(result) != "TWO three!" || string(source) != "one two three" {
		t.Fatalf("result = %q, source = %q", result, source)
	}
}

func TestApplyRejectsInvalidAndOverlappingEdits(t *testing.T) {
	for _, edits := range [][]edit.TextEdit{
		{{Source: offsetSpan(-1, 0)}},
		{{Source: offsetSpan(2, 1)}},
		{{Source: offsetSpan(0, 4)}},
		{{Source: offsetSpan(0, 2)}, {Source: offsetSpan(1, 3)}},
		{{Source: offsetSpan(1, 1)}, {Source: offsetSpan(1, 1)}},
	} {
		if _, err := edit.Apply([]byte("abc"), edits); err == nil {
			t.Errorf("Apply(%+v) returned no error", edits)
		}
	}
}

func TestApplyWithoutEditsReturnsIndependentBytes(t *testing.T) {
	source := []byte("source")
	result, err := edit.Apply(source, nil)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	result[0] = 'S'
	if string(source) != "source" {
		t.Fatalf("result aliases source: %q", source)
	}
}

func offsetSpan(start, end int) ast.Span {
	return ast.Span{Start: ast.Position{Offset: start}, End: ast.Position{Offset: end}}
}
