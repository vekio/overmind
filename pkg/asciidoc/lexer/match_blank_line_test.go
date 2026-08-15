package lexer

import "testing"

func TestMatchBlankLine(t *testing.T) {
	for _, raw := range []string{"", " ", "\t", " \t  \t"} {
		token, matched := matchBlankLine(raw)
		if !matched || token.Kind != LineBlank || token.Raw != raw {
			t.Errorf("matchBlankLine(%q) = (%+v, %t), want blank token", raw, token, matched)
		}
	}
}

func TestMatchBlankLineRejectsContentAndUnicodeWhitespace(t *testing.T) {
	for _, raw := range []string{"text", "  text", "\u00a0"} {
		if _, matched := matchBlankLine(raw); matched {
			t.Errorf("matchBlankLine(%q) matched, want false", raw)
		}
	}
}
