package lexer

import "testing"

func TestMatchThematicBreak(t *testing.T) {
	token, matched := matchThematicBreak("'''")
	if !matched || token.Kind != LineThematicBreak || token.Raw != "'''" {
		t.Fatalf("matchThematicBreak() = (%+v, %t), want thematic break", token, matched)
	}
}

func TestMatchThematicBreakRequiresExactMarker(t *testing.T) {
	for _, raw := range []string{"''", "''''", " '''", "''' ", "---", "***"} {
		if _, matched := matchThematicBreak(raw); matched {
			t.Errorf("matchThematicBreak(%q) matched, want false", raw)
		}
	}
}
