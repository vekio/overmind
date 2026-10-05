package codecs

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/vekio/overmind/internal/ports"
)

func TestRawStampPreservesSpacingDuplicateDeclarationsAndBody(t *testing.T) {
	source := []byte("= Note\r\n:custom: keep\r\n:overmind-updated-at: 2020-01-01T00:00:00Z\r\n:overmind-updated-at:\t 2021-01-01T00:00:00Z  \t\r\n\r\n:overmind-updated-at: body\r\nLast line")
	at := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	updated, err := (RawNoteCodec{}).Stamp(source, at)
	if err != nil {
		t.Fatal(err)
	}
	expected := bytes.Replace(source, []byte("2021-01-01T00:00:00Z"), []byte("2026-10-05T12:00:00Z"), 1)
	if !bytes.Equal(updated, expected) {
		t.Fatalf("changed unrelated source:\n%q", updated)
	}
}

func TestRawDecodeReportsReadableLocation(t *testing.T) {
	source := []byte(":overmind-id: 11111111-1111-4111-8111-111111111111\n:overmind-type: inbox\n:overmind-created-at: 2026-10-05T00:00:00Z\n:overmind-updated-at: 2026-10-05T00:00:00Z\n\n----\nUnclosed listing\n")
	codec := RawNoteCodec{}
	if _, err := codec.Inspect(source); err != nil {
		t.Fatalf("repairable document not inspectable: %v", err)
	}
	_, err := codec.Decode(ports.NoteKindInbox, source)
	if err == nil || !strings.Contains(err.Error(), "line 6, column 1: unclosed listing block") {
		t.Fatalf("missing readable error: %v", err)
	}
}

func TestRawDecodeDistinguishesListingDelimiterFromThreeHyphens(t *testing.T) {
	header := ":overmind-id: 11111111-1111-4111-8111-111111111111\n:overmind-type: inbox\n:overmind-created-at: 2026-10-05T00:00:00Z\n:overmind-updated-at: 2026-10-05T00:00:00Z\n\n"
	for _, tc := range []struct {
		name, body string
		invalid    bool
	}{
		{"three hyphens do not open a listing", "---\nSome text\n", false},
		{"four hyphens require a closing delimiter", "----\nSome text\n", true},
		{"closed listing", "----\nSome text\n----\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := (RawNoteCodec{}).Decode(ports.NoteKindInbox, []byte(header+tc.body))
			if (err != nil) != tc.invalid {
				t.Fatalf("unexpected validation result: %v", err)
			}
		})
	}
}
