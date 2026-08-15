package sqliteindex

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

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
	document := ports.IndexedDocument{
		ID:      id,
		Path:    "knowledge/page.adoc",
		Content: []byte("= Page\n"),
		Attributes: map[string]string{
			domain.AttributeType:  "page",
			domain.AttributeTitle: "Page",
		},
	}
	if err := store.Upsert(context.Background(), document); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := store.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Path != document.Path || string(got.Content) != string(document.Content) || got.Attributes[domain.AttributeTitle] != "Page" {
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
	if err := store.Upsert(context.Background(), ports.IndexedDocument{ID: oldID, Path: "old.adoc", Content: []byte("old")}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if err := store.ReplaceAll(context.Background(), []ports.IndexedDocument{{ID: newID, Path: "new.adoc", Content: []byte("new")}}); err != nil {
		t.Fatalf("ReplaceAll() error = %v", err)
	}

	if _, err := store.GetByID(context.Background(), oldID); !errors.Is(err, ports.ErrIndexedDocumentNotFound) {
		t.Fatalf("GetByID(old) error = %v", err)
	}
	if _, err := store.GetByID(context.Background(), newID); err != nil {
		t.Fatalf("GetByID(new) error = %v", err)
	}
}
