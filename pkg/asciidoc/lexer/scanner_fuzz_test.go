package lexer

import (
	"bytes"
	"testing"
)

func FuzzScannerReconstructsSource(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		[]byte("text"),
		[]byte("one\ntwo\r\nthree\rfour"),
		[]byte("== Título\n* 世界\n"),
		{0xff, '\n', 0x00, '\r'},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, source []byte) {
		scanner := New(bytes.NewReader(source))
		var reconstructed []byte
		previousOffset := 0
		for tokenCount := 0; ; tokenCount++ {
			if tokenCount > len(source)+1 {
				t.Fatal("scanner did not reach EOF")
			}
			token, err := scanner.Next()
			if err != nil {
				t.Fatalf("Next() error = %v", err)
			}
			if token.Kind == LineEOF {
				if token.Source.Start.Offset != len(source) {
					t.Fatalf("EOF offset = %d, want %d", token.Source.Start.Offset, len(source))
				}
				break
			}
			if token.Kind == LineInvalid {
				t.Fatal("scanner emitted an invalid token")
			}
			if token.Source.Start.Offset != previousOffset {
				t.Fatalf("token start offset = %d, want %d", token.Source.Start.Offset, previousOffset)
			}
			reconstructed = append(reconstructed, token.SourceText()...)
			previousOffset += len(token.SourceText())
		}
		if !bytes.Equal(reconstructed, source) {
			t.Fatalf("reconstructed source = %q, want %q", reconstructed, source)
		}
	})
}

func FuzzMatchLinePreservesRaw(f *testing.F) {
	for _, seed := range []string{"", "plain", "====", "term:: description", "image::target[]"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		token := matchLine(raw)
		if token.Kind == LineInvalid || token.Kind == LineEOF {
			t.Fatalf("matchLine(%q) emitted %s", raw, token.Kind)
		}
		if token.Raw != raw {
			t.Fatalf("matchLine(%q) changed raw to %q", raw, token.Raw)
		}
	})
}
