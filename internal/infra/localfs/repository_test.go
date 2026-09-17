package localfs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

func documentID(t *testing.T, value string) domain.DocumentID {
	t.Helper()
	id, err := domain.NewDocumentID(value)
	if err != nil {
		t.Fatalf("NewDocumentID() error = %v", err)
	}
	return id
}

func TestStorePersistsDocumentsByID(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store := New(root)
	id := documentID(t, "page-id")
	blob := ports.Blob{ID: id, Content: []byte("= Page\n")}

	if err := store.Create(ctx, blob); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "page-id.adoc")); err != nil {
		t.Fatalf("stored file error = %v", err)
	}
	got, err := store.Get(ctx, id)
	if err != nil || got.ID != id || string(got.Content) != string(blob.Content) || got.Revision == "" {
		t.Fatalf("Get() = (%+v, %v)", got, err)
	}
	ids, err := store.List(ctx)
	if err != nil || len(ids) != 1 || ids[0] != id {
		t.Fatalf("List() = (%v, %v)", ids, err)
	}
	if err := store.Delete(ctx, id); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := store.Get(ctx, id); !errors.Is(err, ports.ErrBlobNotFound) {
		t.Fatalf("Get() error = %v", err)
	}
}

func TestStoreCreateDoesNotReplaceExistingDocument(t *testing.T) {
	store := New(t.TempDir())
	id := documentID(t, "page-id")
	if err := store.Create(context.Background(), ports.Blob{ID: id, Content: []byte("old")}); err != nil {
		t.Fatal(err)
	}
	if err := store.Create(context.Background(), ports.Blob{ID: id, Content: []byte("new")}); !errors.Is(err, ports.ErrBlobAlreadyExists) {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestStorePutAndUpdateReplaceCompleteDocument(t *testing.T) {
	ctx := context.Background()
	store := New(t.TempDir())
	id := documentID(t, "page-id")
	if err := store.Put(ctx, ports.Blob{ID: id, Content: []byte("old")}); err != nil {
		t.Fatal(err)
	}
	current, _ := store.Get(ctx, id)
	revision, err := store.Update(ctx, ports.Blob{ID: id, Content: []byte("new")}, current.Revision)
	if err != nil || revision == current.Revision {
		t.Fatalf("Update() = (%q, %v)", revision, err)
	}
	updated, _ := store.Get(ctx, id)
	if string(updated.Content) != "new" || updated.Revision != revision {
		t.Fatalf("updated = %+v", updated)
	}
	if _, err := store.Update(ctx, ports.Blob{ID: id, Content: []byte("stale")}, current.Revision); !errors.Is(err, ports.ErrBlobChanged) {
		t.Fatalf("Update(stale) error = %v", err)
	}
}

func TestStoreRejectsZeroIDAndIgnoresUnmanagedFiles(t *testing.T) {
	root := t.TempDir()
	store := New(root)
	if err := store.Put(context.Background(), ports.Blob{}); !errors.Is(err, domain.ErrInvalidDocumentID) {
		t.Fatalf("Put() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "not-a-document.txt"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	ids, err := store.List(context.Background())
	if err != nil || len(ids) != 0 {
		t.Fatalf("List() = (%v, %v)", ids, err)
	}
}
