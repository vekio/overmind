package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/domain"
	"github.com/vekio/overmind/internal/infra/localfs"
	"github.com/vekio/overmind/internal/infra/sqliteindex"
	"github.com/vekio/overmind/internal/ports"
)

func TestDeleteNoteRemovesFileAndIndexedMetadata(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	files := localfs.New(filepath.Join(root, "notes"))
	index, err := sqliteindex.New(ctx, filepath.Join(root, "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = index.Close() })
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	path, err := files.Write(ctx, id, []byte("note"))
	if err != nil {
		t.Fatal(err)
	}
	date := time.Now()
	if err := index.UpsertRecord(ctx, ports.IndexRecord{
		ID: id, Kind: domain.NoteKindPage, CreatedAt: date, UpdatedAt: date,
		Attributes: []ports.IndexAttribute{{Name: "title", Value: "Example"}}, Tags: []string{"tag"},
	}); err != nil {
		t.Fatal(err)
	}

	application := app.New(app.Dependencies{Deleter: files, Index: index})
	if err := application.Commands.DeleteNote.Handle(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("deleted file still exists: %v", err)
	}
	notes, err := index.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 {
		t.Fatalf("deleted note remains indexed: %+v", notes)
	}
	if err := application.Commands.DeleteNote.Handle(ctx, id); err != nil {
		t.Fatalf("deleting an already absent note: %v", err)
	}
}
