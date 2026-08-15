package lexer

import "testing"

func TestMatchLineDispatchesSupportedPrefixes(t *testing.T) {
	tests := []struct {
		raw  string
		kind LineTokenKind
	}{
		{raw: "", kind: LineBlank},
		{raw: " \t", kind: LineBlank},
		{raw: "== Title", kind: LineHeading},
		{raw: "= Document", kind: LineHeading},
		{raw: ":toc: left", kind: LineAttributeEntry},
		{raw: "[source,go]", kind: LineAttributeList},
		{raw: "[[intro]]", kind: LineAnchor},
		{raw: ".Block title", kind: LineBlockTitle},
		{raw: "----", kind: LineDelimiter},
		{raw: "////", kind: LineDelimiter},
		{raw: "* item", kind: LineListItem},
		{raw: "** nested", kind: LineListItem},
		{raw: ". item", kind: LineListItem},
		{raw: ".. nested", kind: LineListItem},
		{raw: "NOTE: Remember this.", kind: LineAdmonition},
		{raw: "// comment", kind: LineComment},
		{raw: "'''", kind: LineThematicBreak},
		{raw: "+", kind: LineContinuation},
		{raw: "<<<", kind: LinePageBreak},
		{raw: "CPU:: processor", kind: LineDescriptionListItem},
		{raw: "image::sunset.jpg[]", kind: LineBlockMacro},
		{raw: "--", kind: LineDelimiter},
		{raw: ":===", kind: LineDelimiter},
		{raw: "plain", kind: LineText},
	}

	for _, test := range tests {
		t.Run(test.kind.String()+"/"+test.raw, func(t *testing.T) {
			token := matchLine(test.raw)
			if token.Kind != test.kind || token.Raw != test.raw {
				t.Fatalf("matchLine(%q) = %s raw %q, want %s preserving raw", test.raw, token.Kind, token.Raw, test.kind)
			}
		})
	}
}

func TestMatchLineFallsBackToText(t *testing.T) {
	for _, raw := range []string{
		" == Indented",
		"==Missing space",
		"[source",
		"---=",
		"[[missing-end",
		":missing-closing",
		"- markdown item",
		"term::description",
	} {
		token := matchLine(raw)
		if token.Kind != LineText || token.Raw != raw {
			t.Errorf("matchLine(%q) = %s raw %q, want TEXT preserving raw", raw, token.Kind, token.Raw)
		}
	}
}
