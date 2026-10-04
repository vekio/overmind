package shared_test

import (
	"errors"
	"github.com/vekio/overmind/internal/domain/shared"
	"testing"
	"time"
)

func TestEntityMetadataRequiresChronologicalTimes(t *testing.T) {
	at := time.Date(2026, 10, 4, 12, 0, 0, 123, time.UTC)
	for _, pair := range [][2]time.Time{{{}, at}, {at, {}}, {at, at.Add(-time.Second)}} {
		if _, err := shared.NewEntityMetadata(pair[0], pair[1]); !errors.Is(err, shared.ErrInvalidEntityMetadata) {
			t.Fatalf("invalid metadata accepted: %v", err)
		}
	}
	original, err := shared.NewEntityMetadata(at, at)
	if err != nil {
		t.Fatal(err)
	}
	later := at.Add(time.Hour)
	updated, err := original.Updated(later)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.CreatedAt().Equal(at) || !updated.UpdatedAt().Equal(later) || !original.UpdatedAt().Equal(at) {
		t.Fatal("metadata update must preserve creation time and original value")
	}
	if _, err := updated.Updated(at); !errors.Is(err, shared.ErrInvalidEntityMetadata) {
		t.Fatal("update moved backwards")
	}
}
