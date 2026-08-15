package lexer

import "testing"

func TestMatchHeading(t *testing.T) {
	tests := []struct {
		raw  string
		want HeadingPayload
	}{
		{raw: "= Document Title", want: HeadingPayload{Level: 0, Title: "Document Title", TitleByteOffset: 2}},
		{raw: "== Title", want: HeadingPayload{Level: 1, Title: "Title", TitleByteOffset: 3}},
		{raw: "====== Deep", want: HeadingPayload{Level: 5, Title: "Deep", TitleByteOffset: 7}},
		{raw: "== \t Título \t", want: HeadingPayload{Level: 1, Title: "Título", TitleByteOffset: 5}},
	}

	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			token, matched := matchHeading(test.raw)
			if !matched || token.Kind != LineHeading || token.Raw != test.raw || token.Heading != test.want {
				t.Fatalf("matchHeading(%q) = (%+v, %t), want payload %+v", test.raw, token, matched, test.want)
			}
		})
	}
}

func TestMatchHeadingRejectsInvalidForms(t *testing.T) {
	for _, raw := range []string{"=Missing", "==Missing", "======= Too deep", "==   ", " == Indented"} {
		if _, matched := matchHeading(raw); matched {
			t.Errorf("matchHeading(%q) matched, want false", raw)
		}
	}
}
