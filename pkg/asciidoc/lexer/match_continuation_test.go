package lexer

import "testing"

func TestMatchContinuation(t *testing.T) {
	token, matched := matchContinuation("+")
	if !matched || token.Kind != LineContinuation || token.Raw != "+" {
		t.Fatalf("matchContinuation() = (%+v, %t), want continuation", token, matched)
	}
}

func TestMatchContinuationRequiresIsolatedMarker(t *testing.T) {
	for _, raw := range []string{"", "++", "+ ", " +"} {
		if _, matched := matchContinuation(raw); matched {
			t.Errorf("matchContinuation(%q) matched, want false", raw)
		}
	}
}
