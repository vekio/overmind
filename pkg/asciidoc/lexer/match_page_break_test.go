package lexer

import "testing"

func TestMatchPageBreak(t *testing.T) {
	token, matched := matchPageBreak("<<<")
	if !matched || token.Kind != LinePageBreak || token.Raw != "<<<" {
		t.Fatalf("matchPageBreak() = (%+v, %t), want page break", token, matched)
	}
}

func TestMatchPageBreakRequiresExactMarker(t *testing.T) {
	for _, raw := range []string{"<<", "<<<<", " <<<", "<<< "} {
		if _, matched := matchPageBreak(raw); matched {
			t.Errorf("matchPageBreak(%q) matched, want false", raw)
		}
	}
}
