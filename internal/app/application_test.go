package app_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"uuid"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/infra/asciidocnote"
	"git.casta.me/alberto/overmind/internal/infra/idgenerator"
	"git.casta.me/alberto/overmind/internal/infra/localfs"
	"git.casta.me/alberto/overmind/internal/infra/renderer"
	"git.casta.me/alberto/overmind/internal/infra/sqliteindex"
)

type appFixture struct {
	application *app.Application
	index       *sqliteindex.Store
	notesPath   string
}

func newAppFixture(t *testing.T) appFixture {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	index, err := sqliteindex.New(ctx, filepath.Join(root, "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = index.Close() })
	documentRenderer, err := renderer.New()
	if err != nil {
		t.Fatal(err)
	}
	notesPath := filepath.Join(root, "notes")
	files := localfs.New(notesPath)
	return appFixture{
		application: app.New(app.Dependencies{
			IDGenerator: idgenerator.New(), Renderer: documentRenderer,
			Writer: files, Reader: files, Deleter: files, Index: index, Lister: index,
			Walker: localfs.NewWalker(notesPath), Parser: asciidocnote.Parser{},
		}),
		index: index, notesPath: notesPath,
	}
}

func TestNoteUseCasesAcrossVaultAndIndex(t *testing.T) {
	ctx := context.Background()
	fixture := newAppFixture(t)
	application := fixture.application
	title, _ := domain.NewTitle("Project plan")
	area, _ := domain.NewArea("Work/Ideas")
	tag, _ := domain.NewTag("Important")
	tags, _ := domain.NewTags(tag)
	page, err := application.Commands.CreatePage.Handle(ctx, app.CreatePageCommand{Title: title, Area: area, Tags: tags})
	if err != nil {
		t.Fatal(err)
	}
	url, _ := domain.NewURL("https://example.com/article")
	bookmark, err := application.Commands.CreateBookmark.Handle(ctx, app.CreateBookmarkCommand{URL: url, Tags: tags})
	if err != nil {
		t.Fatal(err)
	}
	inbox, err := application.Commands.Capture.Handle(ctx, app.CaptureCommand{Content: "quick thought"})
	if err != nil {
		t.Fatal(err)
	}
	journal, err := application.Commands.CreateJournal.Handle(ctx, app.CreateJournalCommand{})
	if err != nil {
		t.Fatal(err)
	}
	journalID, exists, err := application.Queries.FindJournal.Handle(ctx, journal.Journal.Date())
	if err != nil || !exists || journalID != journal.Journal.Metadata().ID() {
		t.Fatalf("find journal = %s, %t, %v", journalID, exists, err)
	}

	listed, err := application.Queries.ListNotes.Handle(ctx)
	if err != nil || len(listed.Notes) != 4 {
		t.Fatalf("list notes = %+v, %v", listed, err)
	}
	pageID := page.Page.Metadata().ID()
	original, err := application.Queries.OpenNote.Handle(ctx, pageID)
	if err != nil {
		t.Fatal(err)
	}
	updated := bytes.Replace(original, []byte("= Project plan"), []byte("= Revised plan"), 1)
	updated = append(updated, []byte("\nA new paragraph.\n")...)
	result, err := application.Commands.UpdateNote.Handle(ctx, app.UpdateNoteCommand{ID: pageID, Original: original, Source: updated})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(result.Source, []byte("= Revised plan")) || !bytes.Contains(result.Source, []byte("A new paragraph.")) {
		t.Fatalf("updated source lost changes: %s", result.Source)
	}
	parser := asciidocnote.Parser{}
	before, err := parser.Parse(original)
	if err != nil {
		t.Fatal(err)
	}
	after, err := parser.Parse(result.Source)
	if err != nil || !after.CreatedAt.Equal(before.CreatedAt) || !after.UpdatedAt.After(before.UpdatedAt) {
		t.Fatalf("updated timestamps = before %+v, after %+v, error %v", before, after, err)
	}
	stored, err := application.Queries.OpenNote.Handle(ctx, pageID)
	if err != nil || !bytes.Equal(stored, result.Source) {
		t.Fatalf("stored source differs from update result: %v", err)
	}
	if _, err := application.Commands.UpdateNote.Handle(ctx, app.UpdateNoteCommand{ID: pageID, Original: original, Source: updated}); !errors.Is(err, app.ErrNoteChanged) {
		t.Fatalf("stale edit = %v, want ErrNoteChanged", err)
	}
	listed, err = application.Queries.ListNotes.Handle(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if note := listedNote(t, listed.Notes, pageID); note.Attributes["title"] != "Revised plan" || note.Attributes["area"] != "work/ideas" {
		t.Fatalf("updated indexed page = %+v", note)
	}
	if note := listedNote(t, listed.Notes, bookmark.Bookmark.Metadata().ID()); note.Attributes["url"] != url.String() {
		t.Fatalf("indexed bookmark = %+v", note)
	}
	if source, err := application.Queries.OpenNote.Handle(ctx, inbox.Inbox.Metadata().ID()); err != nil || !bytes.Contains(source, []byte("quick thought")) {
		t.Fatalf("opened capture = %q, %v", source, err)
	}

	bookmarkID := bookmark.Bookmark.Metadata().ID()
	if err := application.Commands.DeleteNote.Handle(ctx, bookmarkID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(bookmark.Path); !os.IsNotExist(err) {
		t.Fatalf("bookmark file remains: %v", err)
	}
	if err := fixture.index.Delete(ctx, pageID); err != nil {
		t.Fatal(err)
	}
	rebuilt, err := application.Commands.RebuildIndex.Handle(ctx)
	if err != nil || rebuilt.Count != 3 {
		t.Fatalf("rebuild = %+v, %v", rebuilt, err)
	}
	listed, err = application.Queries.ListNotes.Handle(ctx)
	if err != nil || len(listed.Notes) != 3 {
		t.Fatalf("list after rebuild = %+v, %v", listed, err)
	}
	if note := listedNote(t, listed.Notes, pageID); note.Attributes["title"] != "Revised plan" {
		t.Fatalf("rebuilt page = %+v", note)
	}
	if _, err := os.Stat(page.Path); err != nil {
		t.Fatalf("page file missing after rebuild: %v", err)
	}
}

func listedNote(t *testing.T, notes []app.ListedNote, id uuid.UUID) app.ListedNote {
	t.Helper()
	for _, note := range notes {
		if note.ID == id {
			return note
		}
	}
	t.Fatalf("note %s not listed", id)
	return app.ListedNote{}
}
