package sqliteindex

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

func TestStoreUpsertAndGetByID(t *testing.T) {
	store := newTestStore(t)
	document := indexedDocument(t, "page-id", "Page", "knowledge", []string{"go", "ddd"})
	if err := store.Upsert(context.Background(), document); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	got, err := store.GetByID(context.Background(), document.ID)
	if err != nil || got.ID != document.ID || got.Title != "Page" || got.Area != "knowledge" ||
		!reflect.DeepEqual(got.Tags, document.Tags) || got.Attributes["custom"] != "value" {
		t.Fatalf("GetByID() = (%+v, %v)", got, err)
	}
}

func TestStoreReplaceAllRemovesPreviousDocuments(t *testing.T) {
	store := newTestStore(t)
	old := indexedDocument(t, "old-id", "Old", "", nil)
	newDocument := indexedDocument(t, "new-id", "New", "", nil)
	if err := store.Upsert(context.Background(), old); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceAll(context.Background(), []ports.IndexedDocument{newDocument}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetByID(context.Background(), old.ID); !errors.Is(err, ports.ErrIndexedDocumentNotFound) {
		t.Fatalf("GetByID(old) error = %v", err)
	}
	if _, err := store.GetByID(context.Background(), newDocument.ID); err != nil {
		t.Fatalf("GetByID(new) error = %v", err)
	}
}

func TestStoreListsAndFiltersByTitleAreaKindAndTags(t *testing.T) {
	store := newTestStore(t)
	for _, document := range []ports.IndexedDocument{
		indexedDocument(t, "third-id", "Zebra", "archive", []string{"ddd"}),
		indexedDocument(t, "first-id", "Atomic Files", "knowledge/go", []string{"go", "ddd"}),
		indexedDocument(t, "second-id", "Building APIs", "knowledge/api", []string{"go"}),
	} {
		if err := store.Upsert(context.Background(), document); err != nil {
			t.Fatal(err)
		}
	}
	area, _ := domain.NewArea("Knowledge")
	documents, err := store.List(context.Background(), ports.ListIndexedDocumentsFilter{
		Kind: domain.DocumentKindPage, Title: "atomic", Area: area, Tags: documentTags(t, "go", "ddd"),
	})
	if err != nil || len(documents) != 1 || documents[0].ID.String() != "first-id" || documents[0].Area != "knowledge/go" {
		t.Fatalf("List(filtered) = (%+v, %v)", documents, err)
	}

	documents, err = store.List(context.Background(), ports.ListIndexedDocumentsFilter{})
	if err != nil || len(documents) != 3 || documents[0].Title != "Atomic Files" ||
		documents[1].Title != "Building APIs" || documents[2].Title != "Zebra" {
		t.Fatalf("List(ordered) = (%+v, %v)", documents, err)
	}
}

func TestStoreSchemaUsesIDWithoutStoragePath(t *testing.T) {
	store := newTestStore(t)
	rows, err := store.db.QueryContext(context.Background(), "PRAGMA table_info(documents)")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	columns := make(map[string]bool)
	for rows.Next() {
		var columnID, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&columnID, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatal(err)
		}
		columns[name] = true
	}
	if columns["path"] || columns["content"] {
		t.Fatalf("unexpected storage columns: %+v", columns)
	}
	for _, name := range []string{"id", "kind", "title", "area", "created_at", "updated_at"} {
		if !columns[name] {
			t.Fatalf("missing column %q", name)
		}
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := New(context.Background(), filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func indexedDocument(t *testing.T, id, title, area string, tags []string) ports.IndexedDocument {
	t.Helper()
	documentID, err := domain.NewDocumentID(id)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	return ports.IndexedDocument{
		ID: documentID, Kind: domain.DocumentKindPage, Title: title, Area: area, Tags: tags,
		CreatedAt: now, UpdatedAt: now, Attributes: map[string]string{"custom": "value"},
	}
}

func documentTags(t *testing.T, values ...string) domain.Tags {
	t.Helper()
	tags := make([]domain.Tag, len(values))
	for index, value := range values {
		tags[index], _ = domain.NewTag(value)
	}
	result, err := domain.NewTags(tags...)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
