package domain_test

import (
	"errors"
	"testing"
	"time"
	"uuid"

	"git.casta.me/alberto/overmind/internal/domain"
)

func TestJournalRequiresItsFields(t *testing.T) {
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	now := time.Now()
	if _, err := domain.NewJournal(id, domain.Date{}, domain.Tags{}, now); !errors.Is(err, domain.ErrInvalidJournal) {
		t.Fatalf("journal without date = %v", err)
	}
}
