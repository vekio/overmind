package domain_test

import (
	"errors"
	"testing"

	"github.com/vekio/overmind/internal/domain"
)

func TestDateRejectsImpossibleDays(t *testing.T) {
	if _, err := domain.NewDate("2026-02-29"); !errors.Is(err, domain.ErrInvalidDate) {
		t.Fatalf("non-leap February 29 = %v", err)
	}
	date, err := domain.NewDate("2024-02-29")
	if err != nil || date.String() != "2024-02-29" {
		t.Fatalf("leap day = %q, %v", date, err)
	}
}
