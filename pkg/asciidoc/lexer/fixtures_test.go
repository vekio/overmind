package lexer

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestScannerAsciiDocFixtures(t *testing.T) {
	tests := []struct {
		name string
		want []LineToken
	}{
		{
			name: "supported.adoc",
			want: []LineToken{
				{Kind: LineHeading, Raw: "== Introduction", Heading: HeadingPayload{Level: 1, Title: "Introduction", TitleByteOffset: 3}},
				{Kind: LineBlank, Raw: ""},
				{Kind: LineAttributeList, Raw: "[source,go]", AttributeList: AttributeListPayload{Entries: []AttributeListEntry{{Value: "source", ValueByteOffset: 1}, {Value: "go", ValueByteOffset: 8}}}},
				{Kind: LineDelimiter, Raw: "----", Delimiter: DelimiterPayload{Kind: DelimiterListing, Marker: "----"}},
				{Kind: LineText, Raw: "package main"},
				{Kind: LineListItem, Raw: "* item", List: ListPayload{Kind: ListUnordered, Level: 1, Marker: "*", Principal: "item", PrincipalByteOffset: 2}},
				{Kind: LineListItem, Raw: "** nested item", List: ListPayload{Kind: ListUnordered, Level: 2, Marker: "**", Principal: "nested item", PrincipalByteOffset: 3}},
				{Kind: LineListItem, Raw: ". ordered", List: ListPayload{Kind: ListOrdered, Level: 1, Marker: ".", Principal: "ordered", PrincipalByteOffset: 2}},
				{Kind: LineListItem, Raw: ".. nested ordered", List: ListPayload{Kind: ListOrdered, Level: 2, Marker: "..", Principal: "nested ordered", PrincipalByteOffset: 3}},
				{Kind: LineDelimiter, Raw: "----", Delimiter: DelimiterPayload{Kind: DelimiterListing, Marker: "----"}},
			},
		},
		{
			name: "fallback.adoc",
			want: []LineToken{
				{Kind: LineText, Raw: "=Missing title"},
				{Kind: LineText, Raw: "==Missing space"},
				{Kind: LineText, Raw: "======= Too deep"},
				{Kind: LineText, Raw: "[source"},
				{Kind: LineAttributeList, Raw: "[source,go,linenums]", AttributeList: AttributeListPayload{Entries: []AttributeListEntry{{Value: "source", ValueByteOffset: 1}, {Value: "go", ValueByteOffset: 8}, {Value: "linenums", ValueByteOffset: 11}}}},
				{Kind: LineText, Raw: "- markdown item"},
				{Kind: LineText, Raw: "*. mixed markers"},
				{Kind: LineDescriptionListItem, Raw: "term:: description", DescriptionList: DescriptionListPayload{Term: "term", Marker: "::", MarkerByteOffset: 4, Description: "description", DescriptionByteOffset: 7}},
				{Kind: LineDelimiter, Raw: "....", Delimiter: DelimiterPayload{Kind: DelimiterLiteral, Marker: "...."}},
			},
		},
		{
			name: "document_lines.adoc",
			want: []LineToken{
				{Kind: LineHeading, Raw: "= Document Title", Heading: HeadingPayload{Level: 0, Title: "Document Title", TitleByteOffset: 2}},
				{Kind: LineAttributeEntry, Raw: ":toc:", AttributeEntry: AttributeEntryPayload{Operation: AttributeSet, Name: "toc", NameByteOffset: 1, ValueByteOffset: 5}},
				{Kind: LineAttributeEntry, Raw: ":icons: font", AttributeEntry: AttributeEntryPayload{Operation: AttributeSet, Name: "icons", NameByteOffset: 1, Value: "font", ValueByteOffset: 8}},
				{Kind: LineAttributeEntry, Raw: ":!sectids:", AttributeEntry: AttributeEntryPayload{Operation: AttributeUnset, Name: "sectids", NameByteOffset: 2, ValueByteOffset: 10}},
				{Kind: LineAttributeEntry, Raw: ":sectnums!:", AttributeEntry: AttributeEntryPayload{Operation: AttributeUnset, Name: "sectnums", NameByteOffset: 1, ValueByteOffset: 11}},
				{Kind: LineAnchor, Raw: "[[intro]]", Anchor: AnchorPayload{ID: "intro", IDByteOffset: 2}},
				{Kind: LineAnchor, Raw: "[[api,API reference]]", Anchor: AnchorPayload{ID: "api", IDByteOffset: 2, RefText: "API reference", RefTextByteOffset: 6}},
				{Kind: LineBlockTitle, Raw: ".Block title", BlockTitle: BlockTitlePayload{Title: "Block title", TitleByteOffset: 1}},
				{Kind: LineComment, Raw: "// A line comment", Comment: CommentPayload{Content: "A line comment", ContentByteOffset: 3}},
				{Kind: LineComment, Raw: "//", Comment: CommentPayload{ContentByteOffset: 2}},
				{Kind: LineDelimiter, Raw: "////", Delimiter: DelimiterPayload{Kind: DelimiterComment, Marker: "////"}},
				{Kind: LineDelimiter, Raw: "----", Delimiter: DelimiterPayload{Kind: DelimiterListing, Marker: "----"}},
				{Kind: LineDelimiter, Raw: "....", Delimiter: DelimiterPayload{Kind: DelimiterLiteral, Marker: "...."}},
				{Kind: LineDelimiter, Raw: "====", Delimiter: DelimiterPayload{Kind: DelimiterExample, Marker: "===="}},
				{Kind: LineDelimiter, Raw: "****", Delimiter: DelimiterPayload{Kind: DelimiterSidebar, Marker: "****"}},
				{Kind: LineDelimiter, Raw: "____", Delimiter: DelimiterPayload{Kind: DelimiterQuote, Marker: "____"}},
				{Kind: LineDelimiter, Raw: "++++", Delimiter: DelimiterPayload{Kind: DelimiterPassthrough, Marker: "++++"}},
				{Kind: LineDelimiter, Raw: "|===", Delimiter: DelimiterPayload{Kind: DelimiterTable, Marker: "|==="}},
			},
		},
		{
			name: "block_metadata.adoc",
			want: []LineToken{
				{Kind: LineAttributeList, Raw: "[source,go]", AttributeList: AttributeListPayload{Entries: []AttributeListEntry{{Value: "source", ValueByteOffset: 1}, {Value: "go", ValueByteOffset: 8}}}},
				{Kind: LineAttributeList, Raw: "[,ruby]", AttributeList: AttributeListPayload{Entries: []AttributeListEntry{{ValueByteOffset: 1}, {Value: "ruby", ValueByteOffset: 2}}}},
				{Kind: LineAttributeList, Raw: "[#example]", AttributeList: AttributeListPayload{Entries: []AttributeListEntry{{Value: "#example", ValueByteOffset: 1}}}},
				{Kind: LineAttributeList, Raw: `[cols="1,1", options="header"]`, AttributeList: AttributeListPayload{Entries: []AttributeListEntry{{Value: `cols="1,1"`, ValueByteOffset: 1}, {Value: `options="header"`, ValueByteOffset: 13}}}},
				{Kind: LineAdmonition, Raw: "NOTE: Remember this.", Admonition: AdmonitionPayload{Kind: AdmonitionNote, Label: "NOTE", Content: "Remember this.", ContentByteOffset: 6}},
				{Kind: LineAdmonition, Raw: "TIP: Try this.", Admonition: AdmonitionPayload{Kind: AdmonitionTip, Label: "TIP", Content: "Try this.", ContentByteOffset: 5}},
				{Kind: LineAdmonition, Raw: "IMPORTANT: Do not forget.", Admonition: AdmonitionPayload{Kind: AdmonitionImportant, Label: "IMPORTANT", Content: "Do not forget.", ContentByteOffset: 11}},
				{Kind: LineAdmonition, Raw: "CAUTION: Proceed carefully.", Admonition: AdmonitionPayload{Kind: AdmonitionCaution, Label: "CAUTION", Content: "Proceed carefully.", ContentByteOffset: 9}},
				{Kind: LineAdmonition, Raw: "WARNING: Dangerous operation.", Admonition: AdmonitionPayload{Kind: AdmonitionWarning, Label: "WARNING", Content: "Dangerous operation.", ContentByteOffset: 9}},
				{Kind: LineThematicBreak, Raw: "'''"},
			},
		},
		{
			name: "structural_lines.adoc",
			want: []LineToken{
				{Kind: LineContinuation, Raw: "+"},
				{Kind: LinePageBreak, Raw: "<<<"},
				{Kind: LineDescriptionListItem, Raw: "CPU:: The processor", DescriptionList: DescriptionListPayload{Term: "CPU", Marker: "::", MarkerByteOffset: 3, Description: "The processor", DescriptionByteOffset: 6}},
				{Kind: LineDescriptionListItem, Raw: "Linux::: kernel", DescriptionList: DescriptionListPayload{Term: "Linux", Marker: ":::", MarkerByteOffset: 5, Description: "kernel", DescriptionByteOffset: 9}},
				{Kind: LineBlockMacro, Raw: "image::sunset.jpg[Sunset]", BlockMacro: BlockMacroPayload{Name: "image", Target: "sunset.jpg", TargetByteOffset: 7, Attributes: "Sunset", AttributesByteOffset: 18}},
				{Kind: LineBlockMacro, Raw: "include::app.rb[lines=1..3]", BlockMacro: BlockMacroPayload{Name: "include", Target: "app.rb", TargetByteOffset: 9, Attributes: "lines=1..3", AttributesByteOffset: 16}},
				{Kind: LineBlockMacro, Raw: "toc::[]", BlockMacro: BlockMacroPayload{Name: "toc", TargetByteOffset: 5, AttributesByteOffset: 6}},
				{Kind: LineDelimiter, Raw: "--", Delimiter: DelimiterPayload{Kind: DelimiterOpen, Marker: "--"}},
				{Kind: LineDelimiter, Raw: "-----", Delimiter: DelimiterPayload{Kind: DelimiterListing, Marker: "-----"}},
				{Kind: LineDelimiter, Raw: "......", Delimiter: DelimiterPayload{Kind: DelimiterLiteral, Marker: "......"}},
				{Kind: LineDelimiter, Raw: "======", Delimiter: DelimiterPayload{Kind: DelimiterExample, Marker: "======"}},
				{Kind: LineDelimiter, Raw: "******", Delimiter: DelimiterPayload{Kind: DelimiterSidebar, Marker: "******"}},
				{Kind: LineDelimiter, Raw: "______", Delimiter: DelimiterPayload{Kind: DelimiterQuote, Marker: "______"}},
				{Kind: LineDelimiter, Raw: "++++++", Delimiter: DelimiterPayload{Kind: DelimiterPassthrough, Marker: "++++++"}},
				{Kind: LineDelimiter, Raw: "//////", Delimiter: DelimiterPayload{Kind: DelimiterComment, Marker: "//////"}},
				{Kind: LineDelimiter, Raw: "|=====", Delimiter: DelimiterPayload{Kind: DelimiterTable, Marker: "|====="}},
				{Kind: LineDelimiter, Raw: ",===", Delimiter: DelimiterPayload{Kind: DelimiterTable, Marker: ",==="}},
				{Kind: LineDelimiter, Raw: ":===", Delimiter: DelimiterPayload{Kind: DelimiterTable, Marker: ":==="}},
				{Kind: LineDelimiter, Raw: "!===", Delimiter: DelimiterPayload{Kind: DelimiterTable, Marker: "!==="}},
			},
		},
		{
			name: "unicode.adoc",
			want: []LineToken{
				{Kind: LineHeading, Raw: "== Título", Heading: HeadingPayload{Level: 1, Title: "Título", TitleByteOffset: 3}},
				{Kind: LineListItem, Raw: "* 世界", List: ListPayload{Kind: ListUnordered, Level: 1, Marker: "*", Principal: "世界", PrincipalByteOffset: 2}},
				{Kind: LineText, Raw: "Texto con café."},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertFixtureTokens(t, filepath.Join("testdata", test.name), test.want)
		})
	}
}

func assertFixtureTokens(t *testing.T, path string, want []LineToken) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Fatal("fixture must end with LF for deterministic EOF assertions")
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer file.Close()

	scanner := New(file)
	offset := 0
	for index, expected := range want {
		token, err := scanner.Next()
		if err != nil {
			t.Fatalf("Next() token %d: %v", index, err)
		}

		expected.Ending = LineEndingLF
		expected.Source = Span{
			Start: Position{Offset: offset, Line: index + 1, Column: 1},
			End: Position{
				Offset: offset + len(expected.Raw),
				Line:   index + 1,
				Column: utf8.RuneCountInString(expected.Raw) + 1,
			},
		}
		if !reflect.DeepEqual(token, expected) {
			t.Fatalf("token %d = %+v, want %+v", index, token, expected)
		}
		offset += len(expected.Raw) + 1
	}

	eof, err := scanner.Next()
	if err != nil {
		t.Fatalf("Next() EOF: %v", err)
	}
	wantPosition := Position{Offset: len(data), Line: len(want) + 1, Column: 1}
	if !reflect.DeepEqual(eof, eofLineToken(wantPosition)) {
		t.Fatalf("EOF = %+v, want at %s", eof, wantPosition)
	}
}
