package lexer

import "testing"

func TestLineTokenString(t *testing.T) {
	token := headingLineToken("== Title", HeadingPayload{Level: 1, Title: "Title", TitleByteOffset: 3})
	token.Source = Span{
		Start: Position{Line: 1, Column: 1},
		End:   Position{Offset: 8, Line: 1, Column: 9},
	}
	if got, want := token.String(), `HEADING("== Title") [1:1@0, 1:9@8)`; got != want {
		t.Fatalf("LineToken.String() = %q, want %q", got, want)
	}
}

func TestZeroValueLineTokenIsInvalid(t *testing.T) {
	var token LineToken
	if token.Kind != LineInvalid {
		t.Fatalf("zero-value LineToken kind = %s, want INVALID", token.Kind)
	}
}

func TestLineTokenKindString(t *testing.T) {
	tests := []struct {
		kind LineTokenKind
		want string
	}{
		{kind: LineInvalid, want: "INVALID"},
		{kind: LineEOF, want: "EOF"},
		{kind: LineBlank, want: "BLANK"},
		{kind: LineHeading, want: "HEADING"},
		{kind: LineAttributeEntry, want: "ATTRIBUTE_ENTRY"},
		{kind: LineAttributeList, want: "ATTRIBUTE_LIST"},
		{kind: LineAnchor, want: "ANCHOR"},
		{kind: LineBlockTitle, want: "BLOCK_TITLE"},
		{kind: LineDelimiter, want: "DELIMITER"},
		{kind: LineListItem, want: "LIST_ITEM"},
		{kind: LineAdmonition, want: "ADMONITION"},
		{kind: LineComment, want: "COMMENT"},
		{kind: LineThematicBreak, want: "THEMATIC_BREAK"},
		{kind: LineContinuation, want: "CONTINUATION"},
		{kind: LinePageBreak, want: "PAGE_BREAK"},
		{kind: LineDescriptionListItem, want: "DESCRIPTION_LIST_ITEM"},
		{kind: LineBlockMacro, want: "BLOCK_MACRO"},
		{kind: LineText, want: "TEXT"},
		{kind: LineTokenKind(255), want: "UNKNOWN"},
	}

	for _, test := range tests {
		if got := test.kind.String(); got != test.want {
			t.Errorf("LineTokenKind(%d).String() = %q, want %q", test.kind, got, test.want)
		}
	}
}

func TestListKindString(t *testing.T) {
	if got := ListUnknown.String(); got != "UNKNOWN" {
		t.Fatalf("ListUnknown.String() = %q, want UNKNOWN", got)
	}
	if got := ListUnordered.String(); got != "UNORDERED" {
		t.Fatalf("ListUnordered.String() = %q, want UNORDERED", got)
	}
	if got := ListOrdered.String(); got != "ORDERED" {
		t.Fatalf("ListOrdered.String() = %q, want ORDERED", got)
	}
	if got := ListKind(255).String(); got != "UNKNOWN" {
		t.Fatalf("unknown ListKind.String() = %q, want UNKNOWN", got)
	}
}
