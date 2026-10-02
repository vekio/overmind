package sqliteindex_test

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain"
	"github.com/vekio/overmind/internal/infra/sqliteindex"
	"github.com/vekio/overmind/internal/ports"
)

func TestReplaceAllRollsBackOnDuplicateJournalDate(t *testing.T) {
	ctx := context.Background()
	store, err := sqliteindex.New(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	when := time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC)
	previous := ports.IndexRecord{
		ID:   uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Kind: domain.NoteKindPage, CreatedAt: when, UpdatedAt: when,
		Attributes: []ports.IndexAttribute{{Name: "title", Value: "Keep me"}}, Tags: []string{"saved"},
	}
	if err := store.UpsertRecord(ctx, previous); err != nil {
		t.Fatal(err)
	}
	journal := func(id string) ports.IndexRecord {
		return ports.IndexRecord{
			ID: uuid.MustParse(id), Kind: domain.NoteKindJournal,
			CreatedAt: when, UpdatedAt: when,
			Attributes: []ports.IndexAttribute{{Name: "date", Value: "2026-09-28"}},
		}
	}
	if err := store.ReplaceAll(ctx, []ports.IndexRecord{
		journal("22222222-2222-4222-8222-222222222222"),
		journal("33333333-3333-4333-8333-333333333333"),
	}); err == nil {
		t.Fatal("two journals for the same date were indexed")
	}
	notes, err := store.List(ctx)
	if err != nil || len(notes) != 1 || notes[0].ID != previous.ID || len(notes[0].Tags) != 1 {
		t.Fatalf("failed replacement changed index: %+v, %v", notes, err)
	}
}
