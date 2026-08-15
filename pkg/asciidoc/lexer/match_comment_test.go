package lexer

import "testing"

func TestMatchLineComment(t *testing.T) {
	tests := []struct {
		raw     string
		content string
		offset  int
	}{
		{raw: "// A comment", content: "A comment", offset: 3},
		{raw: "//comment", content: "comment", offset: 2},
		{raw: "//", content: "", offset: 2},
		{raw: "//-", content: "-", offset: 2},
	}

	for _, test := range tests {
		token, matched := matchLineComment(test.raw)
		want := CommentPayload{Content: test.content, ContentByteOffset: test.offset}
		if !matched || token.Kind != LineComment || token.Raw != test.raw || token.Comment != want {
			t.Errorf("matchLineComment(%q) = (%+v, %t), want payload %+v", test.raw, token, matched, want)
		}
	}
}

func TestMatchLineCommentRejectsBlockDelimiterAndIndentedForms(t *testing.T) {
	for _, raw := range []string{"/", "///", "////", "/////", " // comment"} {
		if _, matched := matchLineComment(raw); matched {
			t.Errorf("matchLineComment(%q) matched, want false", raw)
		}
	}
}
