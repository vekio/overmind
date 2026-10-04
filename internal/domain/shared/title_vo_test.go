package shared_test

import (
	"errors"
	"testing"

	"github.com/vekio/overmind/internal/domain/shared"
)

func TestTitleNormalizesAndRejectsMultiline(t *testing.T) {
	title, err := shared.NewTitle("  Café notes  ")
	if err != nil || title.String() != "Café notes" || title.Slug() != "cafe-notes" {
		t.Fatalf("title = %+v, %v", title, err)
	}
	if _, err := shared.NewTitle("first\nsecond"); !errors.Is(err, shared.ErrInvalidTitle) {
		t.Fatalf("multiline title = %v", err)
	}
}
