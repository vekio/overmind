package lexer

import "testing"

func TestLineEndingTextAndString(t *testing.T) {
	for _, test := range []struct {
		ending LineEnding
		text   string
		name   string
	}{
		{ending: LineEndingNone, name: "NONE"},
		{ending: LineEndingLF, text: "\n", name: "LF"},
		{ending: LineEndingCRLF, text: "\r\n", name: "CRLF"},
		{ending: LineEndingCR, text: "\r", name: "CR"},
		{ending: LineEnding(255), name: "UNKNOWN"},
	} {
		if got := test.ending.Text(); got != test.text {
			t.Errorf("%d.Text() = %q, want %q", test.ending, got, test.text)
		}
		if got := test.ending.String(); got != test.name {
			t.Errorf("%d.String() = %q, want %q", test.ending, got, test.name)
		}
	}
}
