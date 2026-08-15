package lexer

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"testing/iotest"
)

func TestScannerTokensReconstructExactSource(t *testing.T) {
	input := []byte("first\r\nsecond\rthird\nfourth")
	scanner := New(bytes.NewReader(input))
	var reconstructed strings.Builder
	var endings []LineEnding
	for {
		token, err := scanner.Next()
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if token.Kind == LineEOF {
			break
		}
		reconstructed.WriteString(token.SourceText())
		endings = append(endings, token.Ending)
	}

	if got := reconstructed.String(); got != string(input) {
		t.Fatalf("reconstructed source = %q, want %q", got, input)
	}
	want := []LineEnding{LineEndingCRLF, LineEndingCR, LineEndingLF, LineEndingNone}
	if len(endings) != len(want) {
		t.Fatalf("endings = %v, want %v", endings, want)
	}
	for index := range want {
		if endings[index] != want[index] {
			t.Errorf("ending %d = %s, want %s", index, endings[index], want[index])
		}
	}
}

func TestScannerNextStreamsClassifiedLinesAndPositions(t *testing.T) {
	input := "== Título\r\n[source,go]\r----\n* item\n\nplain"
	scanner := New(iotest.OneByteReader(strings.NewReader(input)))

	tests := []struct {
		kind   LineTokenKind
		raw    string
		ending LineEnding
		source Span
	}{
		{
			kind: LineHeading, raw: "== Título", ending: LineEndingCRLF,
			source: Span{Start: Position{Offset: 0, Line: 1, Column: 1}, End: Position{Offset: 10, Line: 1, Column: 10}},
		},
		{
			kind: LineAttributeList, raw: "[source,go]", ending: LineEndingCR,
			source: Span{Start: Position{Offset: 12, Line: 2, Column: 1}, End: Position{Offset: 23, Line: 2, Column: 12}},
		},
		{
			kind: LineDelimiter, raw: "----", ending: LineEndingLF,
			source: Span{Start: Position{Offset: 24, Line: 3, Column: 1}, End: Position{Offset: 28, Line: 3, Column: 5}},
		},
		{
			kind: LineListItem, raw: "* item", ending: LineEndingLF,
			source: Span{Start: Position{Offset: 29, Line: 4, Column: 1}, End: Position{Offset: 35, Line: 4, Column: 7}},
		},
		{
			kind: LineBlank, raw: "", ending: LineEndingLF,
			source: Span{Start: Position{Offset: 36, Line: 5, Column: 1}, End: Position{Offset: 36, Line: 5, Column: 1}},
		},
		{
			kind: LineText, raw: "plain",
			source: Span{Start: Position{Offset: 37, Line: 6, Column: 1}, End: Position{Offset: 42, Line: 6, Column: 6}},
		},
	}

	for _, test := range tests {
		token, err := scanner.Next()
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if token.Kind != test.kind || token.Raw != test.raw || token.Ending != test.ending || token.Source != test.source {
			t.Fatalf("Next() = %+v, want kind=%s raw=%q source=%s", token, test.kind, test.raw, test.source)
		}
	}

	wantEOF := Span{
		Start: Position{Offset: 42, Line: 6, Column: 6},
		End:   Position{Offset: 42, Line: 6, Column: 6},
	}
	for range 2 {
		token, err := scanner.Next()
		if err != nil || token.Kind != LineEOF || token.Source != wantEOF {
			t.Fatalf("Next() at EOF = (%+v, %v), want stable EOF at %s", token, err, wantEOF)
		}
	}
}

func TestScannerDoesNotCreatePhantomLineAfterFinalEnding(t *testing.T) {
	scanner := New(strings.NewReader("text\n"))
	first, err := scanner.Next()
	if err != nil || first.Kind != LineText {
		t.Fatalf("first Next() = (%+v, %v), want text", first, err)
	}
	eof, err := scanner.Next()
	if err != nil || eof.Kind != LineEOF {
		t.Fatalf("second Next() = (%+v, %v), want EOF", eof, err)
	}
	want := Position{Offset: 5, Line: 2, Column: 1}
	if eof.Source.Start != want {
		t.Fatalf("EOF position = %s, want %s", eof.Source.Start, want)
	}
}

func TestScannerReturnsReadErrorAfterBufferedSource(t *testing.T) {
	scanner := New(iotest.TimeoutReader(strings.NewReader("partial")))
	token, err := scanner.Next()
	if err != nil || token.Kind != LineText || token.Raw != "partial" {
		t.Fatalf("first Next() = (%+v, %v), want partial text", token, err)
	}
	if _, err := scanner.Next(); !errors.Is(err, iotest.ErrTimeout) {
		t.Fatalf("second Next() error = %v, want %v", err, iotest.ErrTimeout)
	}
	eof, err := scanner.Next()
	if err != nil || eof.Kind != LineEOF {
		t.Fatalf("third Next() = (%+v, %v), want EOF", eof, err)
	}
}

func TestScannerRejectsNilReader(t *testing.T) {
	for _, scanner := range []*Scanner{nil, New(nil)} {
		if _, err := scanner.Next(); !errors.Is(err, ErrNilReader) {
			t.Fatalf("Next() error = %v, want ErrNilReader", err)
		}
	}
}

func TestScannerEmptyInputReturnsEOF(t *testing.T) {
	token, err := New(strings.NewReader("")).Next()
	if err != nil || token.Kind != LineEOF {
		t.Fatalf("Next() = (%+v, %v), want EOF", token, err)
	}
	if token.Source.Start != sourceOrigin || token.Source.End != sourceOrigin {
		t.Fatalf("EOF source = %s, want origin", token.Source)
	}
}
