package notestore_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/infra/notestore"
	"github.com/vekio/overmind/internal/ports"
)

func TestScanStopsOnVisitorFailureAndRejectsSymlinkNotes(t *testing.T) {
	root := t.TempDir()
	id := uuid.New()
	path := filepath.Join(root, id.String()+".adoc")
	if err := os.WriteFile(path, []byte("Source bytes\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	store := notestore.New(root)
	failure := errors.New("stop scanning")
	calls := 0
	err := store.Scan(context.Background(), func(note ports.Note) error {
		calls++
		if note.ID != id || note.Path != path || string(note.Content) != "Source bytes\r\n" {
			t.Fatalf("scan changed source or identity: %+v", note)
		}
		return failure
	})
	if !errors.Is(err, failure) || calls != 1 {
		t.Fatalf("visitor failure was ignored: calls=%d, err=%v", calls, err)
	}
	link := filepath.Join(root, uuid.New().String()+".adoc")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if err := store.Scan(context.Background(), func(ports.Note) error { return nil }); err == nil {
		t.Fatal("scan followed a managed note symlink")
	}
}

func TestDeleteRejectsDuplicateIDsWithoutRemovingEitherDocument(t *testing.T) {
	root := t.TempDir()
	id := uuid.New()
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatal(err)
	}
	paths := []string{filepath.Join(root, id.String()+".adoc"), filepath.Join(nested, id.String()+".adoc")}
	for _, path := range paths {
		if err := os.WriteFile(path, []byte("preserve"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	store := notestore.New(root)
	if err := store.Delete(context.Background(), id); err == nil {
		t.Fatal("ambiguous deletion succeeded")
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil || string(content) != "preserve" {
			t.Fatalf("ambiguous deletion altered %s: %v", path, err)
		}
	}
}

func TestScanCancellationStopsBeforeNextDocument(t *testing.T) {
	store := notestore.New(t.TempDir())
	for range 2 {
		if _, err := store.Put(context.Background(), ports.Note{ID: uuid.New(), Content: []byte("preserve")}); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	err := store.Scan(ctx, func(ports.Note) error {
		calls++
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("scan continued after cancellation: calls=%d, err=%v", calls, err)
	}
}
