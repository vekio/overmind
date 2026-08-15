package lexer

import "testing"

func TestMatchAnchor(t *testing.T) {
	for _, test := range []struct {
		raw  string
		want AnchorPayload
	}{
		{raw: "[[intro]]", want: AnchorPayload{ID: "intro", IDByteOffset: 2}},
		{raw: "[[intro,Introduction]]", want: AnchorPayload{ID: "intro", IDByteOffset: 2, RefText: "Introduction", RefTextByteOffset: 8}},
		{raw: "[[ intro , Reference text ]]", want: AnchorPayload{ID: "intro", IDByteOffset: 3, RefText: "Reference text", RefTextByteOffset: 11}},
	} {
		token, matched := matchAnchor(test.raw)
		if !matched || token.Kind != LineAnchor || token.Raw != test.raw || token.Anchor != test.want {
			t.Errorf("matchAnchor(%q) = (%+v, %t), want payload %+v", test.raw, token, matched, test.want)
		}
	}
}

func TestMatchAnchorRejectsInvalidForms(t *testing.T) {
	for _, raw := range []string{"[[]]", "[[id]", "[id]]", "[[id,]]", "[[bad id]]", "[[bad[id]]", "text [[id]]"} {
		if _, matched := matchAnchor(raw); matched {
			t.Errorf("matchAnchor(%q) matched, want false", raw)
		}
	}
}
