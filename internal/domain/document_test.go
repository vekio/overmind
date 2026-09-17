package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewDocumentIDNormalizesAndReturnsValue(t *testing.T) {
	id, err := NewDocumentID(" page-id ")
	if err != nil {
		t.Fatalf("NewDocumentID() error = %v", err)
	}
	if got, want := id.String(), "page-id"; got != want {
		t.Fatalf("DocumentID.String() = %q, want %q", got, want)
	}
}

func TestNewDocumentIDRejectsValuesUnsafeForStorage(t *testing.T) {
	for _, value := range []string{"../page", "page/id", `page\id`, "page.adoc"} {
		if _, err := NewDocumentID(value); !errors.Is(err, ErrInvalidDocumentID) {
			t.Fatalf("NewDocumentID(%q) error = %v", value, err)
		}
	}
}

func TestNewDocumentIDRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"", "   ", "page id", "page\tid", "page\nid"} {
		t.Run(value, func(t *testing.T) {
			if _, err := NewDocumentID(value); err == nil {
				t.Fatalf("NewDocumentID(%q) error = nil", value)
			}
		})
	}
}

func TestDocumentIDMarshalsAsJSONText(t *testing.T) {
	id, _ := NewDocumentID("page-id")

	encoded, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if got, want := string(encoded), `"page-id"`; got != want {
		t.Fatalf("json.Marshal() = %s, want %s", got, want)
	}
}

func TestZeroDocumentIDCannotBeMarshaled(t *testing.T) {
	if _, err := (DocumentID{}).MarshalText(); err == nil {
		t.Fatal("DocumentID{}.MarshalText() error = nil")
	}
}
