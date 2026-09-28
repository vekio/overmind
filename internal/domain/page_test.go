package domain_test

import (
	"errors"
	"testing"
	"time"
	"uuid"

	"git.casta.me/alberto/overmind/internal/domain"
)

func TestPageRequiresItsFields(t *testing.T) {
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	now := time.Now()
	if _, err := domain.NewPage(id, domain.Title{}, domain.Area{}, domain.Tags{}, now); !errors.Is(err, domain.ErrInvalidPage) {
		t.Fatalf("page without title = %v", err)
	}
}
