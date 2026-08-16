package domain

import (
	"errors"
	"testing"
)

func TestNewDocumentKind(t *testing.T) {
	kind, err := NewDocumentKind(" page ")
	if err != nil {
		t.Fatalf("NewDocumentKind() error = %v", err)
	}
	if kind != DocumentKindPage || kind.String() != "page" {
		t.Fatalf("kind = %q", kind)
	}
}

func TestNewDocumentKindRejectsUnknownKinds(t *testing.T) {
	for _, value := range []string{"", "journal", "unknown"} {
		if _, err := NewDocumentKind(value); !errors.Is(err, ErrInvalidDocumentKind) {
			t.Fatalf("NewDocumentKind(%q) error = %v", value, err)
		}
	}
}
