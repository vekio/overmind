package domain_test

import (
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain"
)

func TestInboxRequiresMetadataButAllowsEmptyContent(t *testing.T) {
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	now := time.Now()
	if _, err := domain.NewInbox(uuid.Nil(), "", now); !errors.Is(err, domain.ErrInvalidMetadata) {
		t.Fatalf("nil note ID = %v", err)
	}
	if _, err := domain.NewInbox(id, "", time.Time{}); !errors.Is(err, domain.ErrInvalidMetadata) {
		t.Fatalf("zero creation time = %v", err)
	}
	if _, err := domain.NewInbox(id, "", now); err != nil {
		t.Fatalf("empty editable capture = %v", err)
	}
}
