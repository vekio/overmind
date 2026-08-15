package lexer

import "testing"

func TestMatchAdmonitionParagraph(t *testing.T) {
	tests := []struct {
		raw  string
		kind AdmonitionKind
	}{
		{raw: "NOTE: Remember this.", kind: AdmonitionNote},
		{raw: "TIP: Try this.", kind: AdmonitionTip},
		{raw: "IMPORTANT: Do not forget.", kind: AdmonitionImportant},
		{raw: "CAUTION: Proceed carefully.", kind: AdmonitionCaution},
		{raw: "WARNING: Dangerous operation.", kind: AdmonitionWarning},
	}

	for _, test := range tests {
		t.Run(test.kind.String(), func(t *testing.T) {
			token, matched := matchAdmonitionParagraph(test.raw)
			offset := len(test.kind.String()) + 2
			want := AdmonitionPayload{Kind: test.kind, Label: test.kind.String(), LabelByteOffset: 0, Content: test.raw[offset:], ContentByteOffset: offset}
			if !matched || token.Kind != LineAdmonition || token.Raw != test.raw || token.Admonition != want {
				t.Fatalf("matchAdmonitionParagraph(%q) = (%+v, %t), want payload %+v", test.raw, token, matched, want)
			}
		})
	}
}

func TestMatchAdmonitionParagraphRejectsInvalidForms(t *testing.T) {
	for _, raw := range []string{"note: lowercase", "NOTE:no space", "NOTE:  two spaces", "NOTE:", "NOTE: ", "INFO: unknown"} {
		if _, matched := matchAdmonitionParagraph(raw); matched {
			t.Errorf("matchAdmonitionParagraph(%q) matched, want false", raw)
		}
	}
}

func TestAdmonitionKindString(t *testing.T) {
	if got := AdmonitionUnknown.String(); got != "UNKNOWN" {
		t.Fatalf("AdmonitionUnknown.String() = %q, want UNKNOWN", got)
	}
	if got := AdmonitionKind(255).String(); got != "UNKNOWN" {
		t.Fatalf("unknown AdmonitionKind.String() = %q, want UNKNOWN", got)
	}
}
