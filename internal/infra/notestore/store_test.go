package notestore_test

import (
	"context"
	"errors"
	store "github.com/vekio/overmind/internal/infra/notestore"
	"github.com/vekio/overmind/internal/ports"
	"os"
	"path/filepath"
	"testing"
	"uuid"
)

func TestStorePutGetDelete(t *testing.T) {
	root := filepath.Join(t.TempDir(), "notes")
	notes := store.New(root)
	ctx := context.Background()
	id := uuid.New()
	for _, content := range []string{"first", "updated"} {
		if _, err := notes.Put(ctx, ports.Note{ID: id, Content: []byte(content)}); err != nil {
			t.Fatal(err)
		}
		note, err := notes.Get(ctx, id)
		if err != nil || note.ID != id || string(note.Content) != content {
			t.Fatalf("stored note = %+v, %v", note, err)
		}
		raw, err := os.ReadFile(filepath.Join(root, id.String()+".adoc"))
		if err != nil || string(raw) != content {
			t.Fatal("note must use UUID filename")
		}
	}
	if err := notes.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := notes.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := notes.Get(ctx, id); !errors.Is(err, ports.ErrNoteNotFound) {
		t.Fatalf("missing note = %v", err)
	}
	if _, err := notes.Put(ctx, ports.Note{}); err == nil {
		t.Fatal("nil UUID accepted")
	}
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := notes.Put(ctx, ports.Note{ID: id}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := notes.Get(ctx, id); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := notes.Delete(ctx, id); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
