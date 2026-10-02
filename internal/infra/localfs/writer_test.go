package localfs_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/infra/localfs"
)

func TestWriterRoundTripAndDeleteMissingFile(t *testing.T) {
	ctx := context.Background()
	root := filepath.Join(t.TempDir(), "nested", "notes")
	writer := localfs.New(root)
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	path, err := writer.Write(ctx, id, []byte("first"))
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(root, id.String()+".adoc") {
		t.Fatalf("note path = %q", path)
	}
	if _, err := writer.Write(ctx, id, []byte("second")); err != nil {
		t.Fatal(err)
	}
	content, err := writer.Read(ctx, id)
	if err != nil || string(content) != "second" {
		t.Fatalf("updated file = %q, %v", content, err)
	}
	if err := writer.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := writer.Delete(ctx, id); err != nil {
		t.Fatalf("delete missing file = %v", err)
	}
	if _, err := writer.Read(ctx, id); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("read deleted note = %v", err)
	}
	if _, err := writer.Write(ctx, uuid.Nil(), nil); err == nil {
		t.Fatal("nil UUID accepted")
	}
}
