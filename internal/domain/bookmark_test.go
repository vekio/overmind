package domain_test

import (
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain"
)

func TestBookmarkRequiresItsFields(t *testing.T) {
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	now := time.Now()
	if _, err := domain.NewBookmark(id, domain.URL{}, domain.Tags{}, now); !errors.Is(err, domain.ErrInvalidBookmark) {
		t.Fatalf("bookmark without URL = %v", err)
	}
}
