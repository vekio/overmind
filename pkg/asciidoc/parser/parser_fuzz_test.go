package parser_test

import "testing"

// FuzzParserAlwaysReturnsCompleteSourceBounds protects the parser's recovery
// contract for arbitrary input: parsing must not panic, the document must
// always exist, and a bytes.Reader input must be consumed through its final
// byte even when the syntax is incomplete or unsupported.
func FuzzParserAlwaysReturnsCompleteSourceBounds(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		[]byte("plain text"),
		[]byte("== Section\n\nparagraph"),
		[]byte(".Title\n[[id,Reference]]\n[source,go]\n----\ncode\n----"),
		[]byte("----\nunclosed"),
		[]byte("////\nignored\n////"),
		[]byte("[source,go]\r\n-----\rbody\n-----"),
		[]byte(":toc:\n\n* first **strong**\n** nested\n+\nTIP: Visit https://example.com[Home]\n\nTerm:: <<target,Description>>\n\nimage::diagram.svg[Diagram]"),
		[]byte("[cols=\"2*\",options=\"header\"]\r\n|===\r\n|Name |Value\r\n\r\n|Café |*bold*\r\n|==="),
		{0xff, '\n', '[', 'x', ']'},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, source []byte) {
		result := parse(source)
		if result.Document == nil {
			t.Fatal("parser returned a nil document")
		}
		if result.Document.Source.Start.Offset != 0 || result.Document.Source.Start.Line != 1 || result.Document.Source.Start.Column != 1 {
			t.Fatalf("document start = %s, want source origin", result.Document.Source.Start)
		}
		if result.Document.Source.End.Offset != len(source) {
			t.Fatalf("document end offset = %d, want %d", result.Document.Source.End.Offset, len(source))
		}
	})
}
