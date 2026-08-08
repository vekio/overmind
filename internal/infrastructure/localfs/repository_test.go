package localfs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"git.casta.me/alberto/overmind/internal/ports/blobstore"
)

func TestStoreCreateGetListDelete(t *testing.T) {
	ctx := context.Background()
	store := New(t.TempDir())

	blob := blobstore.Blob{
		Path:    "notes/first.adoc",
		Content: []byte("= First\n"),
	}
	if err := store.Create(ctx, blob); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := store.Get(ctx, "notes//first.adoc")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Path != blob.Path {
		t.Fatalf("Get() Path = %q, want %q", got.Path, blob.Path)
	}
	if string(got.Content) != string(blob.Content) {
		t.Fatalf("Get() Content = %q, want %q", got.Content, blob.Content)
	}

	paths, err := store.List(ctx, blobstore.Filter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if want := []string{"notes/first.adoc"}; !reflect.DeepEqual(paths, want) {
		t.Fatalf("List() = %v, want %v", paths, want)
	}

	if err := store.Delete(ctx, blob.Path); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := store.Get(ctx, blob.Path); !errors.Is(err, blobstore.ErrNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, blobstore.ErrNotFound)
	}
}

func TestStoreCreateFailsWhenBlobAlreadyExists(t *testing.T) {
	ctx := context.Background()
	store := New(t.TempDir())

	blob := blobstore.Blob{Path: "note.adoc", Content: []byte("old")}
	if err := store.Create(ctx, blob); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := store.Create(ctx, blobstore.Blob{Path: "note.adoc", Content: []byte("new")}); !errors.Is(err, blobstore.ErrAlreadyExists) {
		t.Fatalf("Create() error = %v, want %v", err, blobstore.ErrAlreadyExists)
	}
}

func TestStorePutReplacesExistingBlob(t *testing.T) {
	ctx := context.Background()
	store := New(t.TempDir())

	if err := store.Put(ctx, blobstore.Blob{Path: "note.adoc", Content: []byte("old")}); err != nil {
		t.Fatalf("Put(old) error = %v", err)
	}
	if err := store.Put(ctx, blobstore.Blob{Path: "note.adoc", Content: []byte("new")}); err != nil {
		t.Fatalf("Put(new) error = %v", err)
	}

	got, err := store.Get(ctx, "note.adoc")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(got.Content) != "new" {
		t.Fatalf("Get() Content = %q, want %q", got.Content, "new")
	}
}

func TestStoreAllowsEmptyBlobContent(t *testing.T) {
	ctx := context.Background()
	store := New(t.TempDir())

	if err := store.Create(ctx, blobstore.Blob{Path: "empty.adoc"}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := store.Get(ctx, "empty.adoc")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if len(got.Content) != 0 {
		t.Fatalf("Get() Content length = %d, want 0", len(got.Content))
	}
}

func TestStoreRejectsInvalidBlobPaths(t *testing.T) {
	ctx := context.Background()
	store := New(t.TempDir())

	for _, blobPath := range []string{
		"",
		"/absolute.adoc",
		"../outside.adoc",
		"notes/../../outside.adoc",
		`notes\windows.adoc`,
	} {
		t.Run(blobPath, func(t *testing.T) {
			err := store.Put(ctx, blobstore.Blob{Path: blobPath, Content: []byte("content")})
			if !errors.Is(err, blobstore.ErrInvalidPath) {
				t.Fatalf("Put() error = %v, want %v", err, blobstore.ErrInvalidPath)
			}
		})
	}
}

func TestStoreDoesNotFollowSymlinksOutsideRoot(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	outside := t.TempDir()
	store := New(root)

	outsidePath := filepath.Join(outside, "outside.adoc")
	if err := os.WriteFile(outsidePath, []byte("outside"), 0o644); err != nil {
		t.Fatalf("os.WriteFile(outside) error = %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatalf("os.Symlink() error = %v", err)
	}

	if _, err := store.Get(ctx, "escape/outside.adoc"); err == nil {
		t.Fatal("Get() error = nil, want root escape error")
	}
	if err := store.Put(ctx, blobstore.Blob{Path: "escape/outside.adoc", Content: []byte("changed")}); err == nil {
		t.Fatal("Put() error = nil, want root escape error")
	}

	content, err := os.ReadFile(outsidePath)
	if err != nil {
		t.Fatalf("os.ReadFile(outside) error = %v", err)
	}
	if got, want := string(content), "outside"; got != want {
		t.Fatalf("outside content = %q, want %q", got, want)
	}

	paths, err := store.List(ctx, blobstore.Filter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("List() = %v, want no symlink entries", paths)
	}
}

func TestStoreListIgnoresGitAndReturnsAllBlobPaths(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store := New(root)

	if err := store.Put(ctx, blobstore.Blob{Path: "b.adoc", Content: []byte("b")}); err != nil {
		t.Fatalf("Put(b) error = %v", err)
	}
	if err := store.Put(ctx, blobstore.Blob{Path: "a.adoc", Content: []byte("a")}); err != nil {
		t.Fatalf("Put(a) error = %v", err)
	}
	if err := store.Put(ctx, blobstore.Blob{Path: "ignored.txt", Content: []byte("ignored")}); err != nil {
		t.Fatalf("Put(ignored) error = %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("os.Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "config.adoc"), []byte("ignored"), 0o644); err != nil {
		t.Fatalf("os.WriteFile(.git) error = %v", err)
	}

	paths, err := store.List(ctx, blobstore.Filter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if want := []string{"a.adoc", "b.adoc", "ignored.txt"}; !reflect.DeepEqual(paths, want) {
		t.Fatalf("List() = %v, want %v", paths, want)
	}
}

func TestStoreListFiltersByPrefixAndSuffix(t *testing.T) {
	ctx := context.Background()
	store := New(t.TempDir())

	for _, blob := range []blobstore.Blob{
		{Path: "notes/a.adoc", Content: []byte("a")},
		{Path: "notes/b.txt", Content: []byte("b")},
		{Path: "archive/c.adoc", Content: []byte("c")},
	} {
		if err := store.Put(ctx, blob); err != nil {
			t.Fatalf("Put(%s) error = %v", blob.Path, err)
		}
	}

	paths, err := store.List(ctx, blobstore.Filter{Prefix: "notes/", Suffix: ".adoc"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if want := []string{"notes/a.adoc"}; !reflect.DeepEqual(paths, want) {
		t.Fatalf("List() = %v, want %v", paths, want)
	}
}
