package domain_test

import (
	"errors"
	"testing"

	"github.com/vekio/overmind/internal/domain"
)

func TestTitleNormalizesAndRejectsMultiline(t *testing.T) {
	title, err := domain.NewTitle("  Café notes  ")
	if err != nil || title.String() != "Café notes" || title.Slug() != "cafe-notes" {
		t.Fatalf("title = %+v, %v", title, err)
	}
	if _, err := domain.NewTitle("first\nsecond"); !errors.Is(err, domain.ErrInvalidTitle) {
		t.Fatalf("multiline title = %v", err)
	}
}
