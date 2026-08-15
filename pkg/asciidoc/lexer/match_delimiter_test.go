package lexer

import "testing"

func TestMatchDelimiter(t *testing.T) {
	tests := []struct {
		raw  string
		kind DelimiterKind
	}{
		{raw: "--", kind: DelimiterOpen},
		{raw: "----", kind: DelimiterListing},
		{raw: "------", kind: DelimiterListing},
		{raw: "....", kind: DelimiterLiteral},
		{raw: "====", kind: DelimiterExample},
		{raw: "****", kind: DelimiterSidebar},
		{raw: "____", kind: DelimiterQuote},
		{raw: "++++", kind: DelimiterPassthrough},
		{raw: "////", kind: DelimiterComment},
		{raw: "|===", kind: DelimiterTable},
		{raw: ",===", kind: DelimiterTable},
		{raw: ":===", kind: DelimiterTable},
		{raw: "!===", kind: DelimiterTable},
		{raw: "|=====", kind: DelimiterTable},
	}

	for _, test := range tests {
		t.Run(test.kind.String(), func(t *testing.T) {
			token, matched := matchDelimiter(test.raw)
			want := DelimiterPayload{Kind: test.kind, Marker: test.raw, MarkerByteOffset: 0}
			if !matched || token.Kind != LineDelimiter || token.Raw != test.raw || token.Delimiter != want {
				t.Fatalf("matchDelimiter(%q) = (%+v, %t), want payload %+v", test.raw, token, matched, want)
			}
		})
	}
}

func TestMatchDelimiterRequiresExactFence(t *testing.T) {
	for _, raw := range []string{"-", "---", "---=", " ----", "---- ", "...", "===", "***=", "|||="} {
		if _, matched := matchDelimiter(raw); matched {
			t.Errorf("matchDelimiter(%q) matched, want false", raw)
		}
	}
}

func TestDelimiterKindString(t *testing.T) {
	if got := DelimiterUnknown.String(); got != "UNKNOWN" {
		t.Fatalf("DelimiterUnknown.String() = %q, want UNKNOWN", got)
	}
	if got := DelimiterKind(255).String(); got != "UNKNOWN" {
		t.Fatalf("unknown DelimiterKind.String() = %q, want UNKNOWN", got)
	}
}
