package documentparser

import (
	"errors"
	"strings"
	"testing"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
)

func TestParseExtractsManagedMetadata(t *testing.T) {
	document, managed, err := Parse(append(validDocumentSource(), []byte(":unrelated: ignored\n\nBody\n")...))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	wantCreatedAt := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	if !managed || document.ID.String() != "page-id" || document.Kind != domain.DocumentKindPage ||
		document.Title.String() != "Page" || document.Tags.Len() != 2 ||
		document.Tags.Strings()[0] != "go" || document.Tags.Strings()[1] != "diseno-de-dominio" ||
		!document.CreatedAt.Equal(wantCreatedAt) || document.Attributes["area"] != "knowledge" {
		t.Fatalf("managed = %t, document = %+v", managed, document)
	}
	if !document.UpdatedAt.Equal(wantCreatedAt) {
		t.Fatalf("UpdatedAt = %v, want %v", document.UpdatedAt, wantCreatedAt)
	}
	if _, exists := document.Attributes["unrelated"]; exists {
		t.Fatalf("unrelated attribute was parsed: %+v", document.Attributes)
	}
}

func TestParseIgnoresMalformedUnmanagedDocument(t *testing.T) {
	document, managed, err := Parse([]byte("= External\n:broken attribute\n"))
	if err != nil || managed || document.ID.String() != "" {
		t.Fatalf("Parse() = (%+v, %t, %v)", document, managed, err)
	}
}

func TestParseValidatesManagedMetadata(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		cause   error
		message string
	}{
		{name: "AsciiDoc", source: string(validDocumentSource()) + "\nBody\n\n= Late title\n", message: "invalid AsciiDoc"},
		{name: "id", source: "= Page\n:overmind-id: invalid id\n:overmind-type: page\n:overmind-created-at: 2026-08-16T10:00:00Z\n\n", cause: domain.ErrInvalidDocumentID},
		{name: "kind", source: "= Page\n:overmind-id: page-id\n:overmind-type: journal\n:overmind-created-at: 2026-08-16T10:00:00Z\n\n", cause: domain.ErrInvalidDocumentKind},
		{name: "title", source: ":overmind-id: page-id\n:overmind-type: page\n:overmind-created-at: 2026-08-16T10:00:00Z\n\nBody\n", cause: domain.ErrInvalidTitle},
		{name: "tag", source: "= Page\n:overmind-id: page-id\n:overmind-type: page\n:overmind-tags: ---\n:overmind-created-at: 2026-08-16T10:00:00Z\n\n", cause: domain.ErrInvalidTag},
		{name: "duplicate tag", source: "= Page\n:overmind-id: page-id\n:overmind-type: page\n:overmind-tags: Go, go\n:overmind-created-at: 2026-08-16T10:00:00Z\n\n", cause: domain.ErrDuplicateTag},
		{name: "created at", source: "= Page\n:overmind-id: page-id\n:overmind-type: page\n:overmind-created-at: yesterday\n\n", message: "invalid overmind-created-at"},
		{name: "updated at", source: "= Page\n:overmind-id: page-id\n:overmind-type: page\n:overmind-created-at: 2026-08-16T10:00:00Z\n:overmind-updated-at: yesterday\n\n", message: "invalid overmind-updated-at"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, managed, err := Parse([]byte(test.source))
			if !managed || err == nil {
				t.Fatalf("Parse() managed = %t, error = %v", managed, err)
			}
			if test.cause != nil && !errors.Is(err, test.cause) {
				t.Fatalf("Parse() error = %v, want %v", err, test.cause)
			}
			if test.message != "" && !strings.Contains(err.Error(), test.message) {
				t.Fatalf("Parse() error = %v", err)
			}
		})
	}
}

func TestTagsFromHeader(t *testing.T) {
	tags, err := tagsFromHeader(" Go, Diseño de dominio ")
	if err != nil {
		t.Fatalf("tagsFromHeader() error = %v", err)
	}
	if got := tags.Strings(); len(got) != 2 || got[0] != "go" || got[1] != "diseno-de-dominio" {
		t.Fatalf("tagsFromHeader() = %v", got)
	}
	empty, err := tagsFromHeader("  ")
	if err != nil || !empty.IsEmpty() {
		t.Fatalf("tagsFromHeader(empty) = (%v, %v)", empty, err)
	}
}

func validDocumentSource() []byte {
	return []byte("= Page\n" +
		":overmind-id: page-id\n" +
		":overmind-type: page\n" +
		":overmind-area: knowledge\n" +
		":overmind-tags: Go, Diseño de dominio\n" +
		":overmind-created-at: 2026-08-16T10:00:00Z\n" +
		":overmind-updated-at: 2026-08-16T10:00:00Z\n")
}
