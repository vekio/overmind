package sqliteindex

import (
	"context"
	"reflect"
	"testing"
	"time"
	"uuid"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

func TestListReturnsIndexedMetadataNewestFirst(t *testing.T) {
	ctx := context.Background()
	store, err := New(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	created := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	older := ports.IndexRecord{
		ID: uuid.MustParse("11111111-1111-4111-8111-111111111111"), Kind: domain.NoteKindPage,
		CreatedAt: created, UpdatedAt: created,
		Attributes: []ports.IndexAttribute{{Name: "title", Value: "Older page"}, {Name: "area", Value: "work"}},
		Tags:       []string{"one", "two"},
	}
	newer := ports.IndexRecord{
		ID: uuid.MustParse("22222222-2222-4222-8222-222222222222"), Kind: domain.NoteKindBookmark,
		CreatedAt: created, UpdatedAt: created.Add(time.Hour),
		Attributes: []ports.IndexAttribute{{Name: "url", Value: "https://example.com"}},
		Tags:       []string{"saved"},
	}
	if err := store.ReplaceAll(ctx, []ports.IndexRecord{older, newer}); err != nil {
		t.Fatal(err)
	}
	got, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != newer.ID || got[1].ID != older.ID {
		t.Fatalf("wrong note order: %#v", got)
	}
	if !reflect.DeepEqual(got[0].Attributes, newer.Attributes) || !reflect.DeepEqual(got[0].Tags, newer.Tags) {
		t.Errorf("newer note metadata = %#v", got[0])
	}
	if !reflect.DeepEqual(got[1].Attributes, []ports.IndexAttribute{{Name: "area", Value: "work"}, {Name: "title", Value: "Older page"}}) || !reflect.DeepEqual(got[1].Tags, older.Tags) {
		t.Errorf("older note metadata = %#v", got[1])
	}
}
