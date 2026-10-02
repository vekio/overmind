package app_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/domain"
)

func TestRebuildRejectsInvalidAndDuplicateNotesWithoutReplacingIndex(t *testing.T) {
	ctx := context.Background()
	fixture := newAppFixture(t)
	title, _ := domain.NewTitle("Kept page")
	page, err := fixture.application.Commands.CreatePage.Handle(ctx, app.CreatePageCommand{Title: title})
	if err != nil {
		t.Fatal(err)
	}
	invalidPath := filepath.Join(fixture.notesPath, "invalid.adoc")
	if err := os.WriteFile(invalidPath, []byte("not an Overmind note\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.application.Commands.RebuildIndex.Handle(ctx); err == nil || !strings.Contains(err.Error(), "invalid.adoc") {
		t.Fatalf("invalid note rebuild = %v", err)
	}
	if err := os.Remove(invalidPath); err != nil {
		t.Fatal(err)
	}
	duplicate, err := os.ReadFile(page.Path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.notesPath, "duplicate.adoc"), duplicate, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.application.Commands.RebuildIndex.Handle(ctx); err == nil || !strings.Contains(err.Error(), "duplicate note ID") {
		t.Fatalf("duplicate note rebuild = %v", err)
	}
	notes, err := fixture.application.Queries.ListNotes.Handle(ctx)
	if err != nil || len(notes.Notes) != 1 || notes.Notes[0].ID != page.Page.Metadata().ID() {
		t.Fatalf("failed rebuild replaced index: %+v, %v", notes, err)
	}
}
