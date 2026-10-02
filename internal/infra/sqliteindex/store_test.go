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

func TestDeleteCascadesAttributesAndTags(t *testing.T) {
	ctx := context.Background()
	store, err := sqliteindex.New(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	when := time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC)
	if err := store.UpsertRecord(ctx, ports.IndexRecord{
		ID: id, Kind: domain.NoteKindPage, CreatedAt: when, UpdatedAt: when,
		Attributes: []ports.IndexAttribute{{Name: "title", Value: "Example"}}, Tags: []string{"saved"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	notes, err := store.List(ctx)
	if err != nil || len(notes) != 0 {
		t.Fatalf("delete left indexed note or dependent rows: %+v, %v", notes, err)
	}
}
