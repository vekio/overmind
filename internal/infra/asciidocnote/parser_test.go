package asciidocnote_test

import (
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/infra/asciidocnote"
)

const pageSource = `= A page
:overmind-id: 11111111-1111-4111-8111-111111111111
:overmind-type: page
:overmind-area: work/ideas
:overmind-tags: alpha, beta
:overmind-created-at: 2026-09-28T10:00:00Z
:overmind-updated-at: 2026-09-28T11:00:00Z

Content.
`

func TestParserRecoversPageMetadata(t *testing.T) {
	record, err := (asciidocnote.Parser{}).Parse([]byte(pageSource))
	if err != nil {
		t.Fatal(err)
	}
	if record.Kind != domain.NoteKindPage || record.CreatedAt.Hour() != 10 || record.UpdatedAt.Hour() != 11 {
		t.Fatalf("page metadata = %+v", record)
	}
	if len(record.Attributes) != 2 || record.Attributes[0].Value != "A page" || record.Attributes[1].Value != "work/ideas" {
		t.Fatalf("page attributes = %+v", record.Attributes)
	}
	if len(record.Tags) != 2 || record.Tags[0] != "alpha" || record.Tags[1] != "beta" {
		t.Fatalf("page tags = %v", record.Tags)
	}
}

func TestParserRejectsMalformedIndexMetadata(t *testing.T) {
	cases := []struct {
		name string
		old  string
		new  string
		want string
	}{
		{"missing ID", ":overmind-id: 11111111-1111-4111-8111-111111111111\n", "", "overmind-id"},
		{"invalid ID", "11111111-1111-4111-8111-111111111111", "not-a-uuid", "invalid overmind-id"},
		{"unknown kind", ":overmind-type: page", ":overmind-type: unknown", "invalid note kind"},
		{"missing page title", "= A page\n", "", "page title is required"},
		{"invalid timestamp", "2026-09-28T11:00:00Z", "yesterday", "invalid overmind-updated-at"},
		{"backwards timestamp", "2026-09-28T11:00:00Z", "2026-09-28T09:00:00Z", "precedes"},
		{"duplicate tags", "alpha, beta", "alpha, alpha", "duplicate tag"},
		{"invalid area", "work/ideas", "work//ideas", "invalid area"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := strings.Replace(pageSource, tc.old, tc.new, 1)
			if source == pageSource {
				t.Fatal("test did not change source")
			}
			_, err := (asciidocnote.Parser{}).Parse([]byte(source))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Parse() error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestParserAllowsUntitledBookmarkAndInbox(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind domain.NoteKind
		attr string
	}{
		{"bookmark", domain.NoteKindBookmark, ":overmind-url: https://example.com\n"},
		{"inbox", domain.NoteKindInbox, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := ":overmind-id: 11111111-1111-4111-8111-111111111111\n" +
				":overmind-type: " + tc.kind.String() + "\n" + tc.attr +
				":overmind-created-at: 2026-09-28T10:00:00Z\n" +
				":overmind-updated-at: 2026-09-28T10:00:00Z\n\n"
			record, err := (asciidocnote.Parser{}).Parse([]byte(source))
			if err != nil || record.Kind != tc.kind {
				t.Fatalf("Parse() = %+v, %v", record, err)
			}
		})
	}
}
