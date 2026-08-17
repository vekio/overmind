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
	store, err := New(context.Background(), filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.Close()

	id, _ := domain.NewDocumentID("page-id")
	createdAt := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	document := ports.IndexedDocument{
		ID:        id,
		Path:      "page/knowledge/page.adoc",
		Kind:      "page",
		Title:     "Page",
		Tags:      []string{"go", "ddd"},
		CreatedAt: createdAt,
		UpdatedAt: createdAt.Add(time.Hour),
		Attributes: map[string]string{
			"area": "knowledge",
		},
	}
	if err := store.Upsert(context.Background(), document); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := store.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Path != document.Path || got.Kind != "page" || got.Title != "Page" || len(got.Tags) != 2 || got.Tags[0] != "go" || got.Tags[1] != "ddd" || !got.CreatedAt.Equal(createdAt) || !got.UpdatedAt.Equal(document.UpdatedAt) || got.Attributes["area"] != "knowledge" {
		t.Fatalf("GetByID() = %+v", got)
	}
}

func TestStoreReplaceAllRemovesPreviousDocuments(t *testing.T) {
	store, err := New(context.Background(), filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.Close()

	oldID, _ := domain.NewDocumentID("old-id")
	newID, _ := domain.NewDocumentID("new-id")
	createdAt := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	if err := store.Upsert(context.Background(), ports.IndexedDocument{
		ID: oldID, Path: "old.adoc", Kind: domain.DocumentKindPage, Title: "Old", CreatedAt: createdAt, UpdatedAt: createdAt,
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if err := store.ReplaceAll(context.Background(), []ports.IndexedDocument{{
		ID: newID, Path: "new.adoc", Kind: domain.DocumentKindPage, Title: "New", CreatedAt: createdAt, UpdatedAt: createdAt,
	}}); err != nil {
		t.Fatalf("ReplaceAll() error = %v", err)
	}

	if _, err := store.GetByID(context.Background(), oldID); !errors.Is(err, ports.ErrIndexedDocumentNotFound) {
		t.Fatalf("GetByID(old) error = %v", err)
	}
	if _, err := store.GetByID(context.Background(), newID); err != nil {
		t.Fatalf("GetByID(new) error = %v", err)
	}
}

func TestStoreListsDocumentsByKindAndAllTags(t *testing.T) {
	store, err := New(context.Background(), filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.Close()

	createdAt := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	for _, document := range []ports.IndexedDocument{
		indexedDocument(t, "third-id", "page/c.adoc", []string{"ddd"}, createdAt),
		indexedDocument(t, "first-id", "page/a.adoc", []string{"go", "ddd"}, createdAt),
		indexedDocument(t, "second-id", "page/b.adoc", []string{"go"}, createdAt),
	} {
		if err := store.Upsert(context.Background(), document); err != nil {
			t.Fatalf("Upsert() error = %v", err)
		}
	}

	documents, err := store.List(context.Background(), ports.ListIndexedDocumentsFilter{
		Kind: domain.DocumentKindPage,
		Tags: documentTags(t, "go", "ddd"),
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(documents) != 1 || documents[0].Path != "page/a.adoc" ||
		!reflect.DeepEqual(documents[0].Tags, []string{"go", "ddd"}) {
		t.Fatalf("List() = %+v", documents)
	}

	documents, err = store.List(context.Background(), ports.ListIndexedDocumentsFilter{Tags: documentTags(t, "go")})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(documents) != 2 || documents[0].Path != "page/a.adoc" || documents[1].Path != "page/b.adoc" {
		t.Fatalf("List() = %+v", documents)
	}

	documents, err = store.List(context.Background(), ports.ListIndexedDocumentsFilter{PathPrefix: "page/b"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(documents) != 1 || documents[0].Path != "page/b.adoc" {
		t.Fatalf("List() = %+v", documents)
	}
}

func indexedDocument(t *testing.T, id, path string, tags []string, createdAt time.Time) ports.IndexedDocument {
	t.Helper()
	documentID, err := domain.NewDocumentID(id)
	if err != nil {
		t.Fatalf("NewDocumentID() error = %v", err)
	}
	return ports.IndexedDocument{
		ID: documentID, Path: path, Kind: domain.DocumentKindPage, Title: path, Tags: tags,
		CreatedAt: createdAt, UpdatedAt: createdAt,
	}
}

func documentTags(t *testing.T, values ...string) domain.Tags {
	t.Helper()
	tags := make([]domain.Tag, len(values))
	for index, value := range values {
		tag, err := domain.NewTag(value)
		if err != nil {
			t.Fatalf("NewTag() error = %v", err)
		}
		tags[index] = tag
	}
	result, err := domain.NewTags(tags...)
	if err != nil {
		t.Fatalf("NewTags() error = %v", err)
	}
	return result
}

func TestStoreSchemaDoesNotContainDocumentContent(t *testing.T) {
	store, err := New(context.Background(), filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.Close()

	rows, err := store.db.QueryContext(context.Background(), "PRAGMA table_info(documents)")
	if err != nil {
		t.Fatalf("PRAGMA table_info() error = %v", err)
	}
	defer rows.Close()
	columns := make(map[string]bool)
	for rows.Next() {
		var columnID, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&columnID, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("Scan() error = %v", err)
		}
		if name == "content" {
			t.Fatal("documents table still contains content column")
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows error = %v", err)
	}
	for _, name := range []string{"id", "path", "kind", "title", "created_at", "updated_at"} {
		if !columns[name] {
			t.Fatalf("documents table does not contain %q", name)
		}
	}
}
