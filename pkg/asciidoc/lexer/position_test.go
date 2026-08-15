package lexer

import "testing"

func TestPositionAndSpanString(t *testing.T) {
	start := Position{Offset: 4, Line: 2, Column: 3}
	end := Position{Offset: 9, Line: 2, Column: 8}
	if got, want := start.String(), "2:3@4"; got != want {
		t.Fatalf("Position.String() = %q, want %q", got, want)
	}
	if got, want := (Span{Start: start, End: end}).String(), "[2:3@4, 2:8@9)"; got != want {
		t.Fatalf("Span.String() = %q, want %q", got, want)
	}
}
