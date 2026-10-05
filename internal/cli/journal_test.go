package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/app/journal"
	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/infra/codecs"
	"github.com/vekio/overmind/internal/ports"
)

func TestJournalCommandPassesTagsAndPrintsID(t *testing.T) {
	client := &otherCommandsClient{}
	got := runCLICommand(t, newJournalCommand(fixedClient(client)), []string{"journal", "--tag", "Daily"}, "")
	if got != "Journal 2026-10-04\nID: 11111111-1111-4111-8111-111111111111\n" || !reflect.DeepEqual(client.journalTags, []string{"Daily"}) {
		t.Fatalf("journal output=%q tags=%v", got, client.journalTags)
	}
}

func TestRootJournalCommandPersistsDateAndRejectsDuplicates(t *testing.T) {
	fixture := newCommandFixture(t)
	date := "2024-02-29"
	note := fixture.create("journal", []string{"journal", "--date", date, "--tag", "Reading", "--tag", "Work"}, "", []string{"reading", "work"})
	var storedDate string
	if err := fixture.db().QueryRow("SELECT date FROM journals WHERE note_id=?", note.ID.String()).Scan(&storedDate); err != nil || storedDate != date {
		t.Fatalf("journal projection=%q err=%v", storedDate, err)
	}
	if !bytes.HasPrefix(note.Source, []byte("= Febrero 29, 2024\n")) {
		t.Fatalf("journal title=%s", note.Source)
	}
	if _, err := fixture.run([]string{"journal", "--date", date, "--tag", "Different"}, ""); !errors.Is(err, journal.ErrAlreadyExists) {
		t.Fatalf("duplicate date=%v", err)
	}
	current, err := os.ReadFile(note.Path)
	if err != nil || !bytes.Equal(current, note.Source) {
		t.Fatal("duplicate journal replaced the original document")
	}
	if _, err := fixture.run([]string{"journal", "--date", "2026-02-29"}, ""); !errors.Is(err, calendar.ErrInvalidDate) {
		t.Fatalf("invalid date=%v", err)
	}
	fixture.requireSingleNote()
}

func TestJournalContentPersistsAndUpdatesWithoutChangingIdentity(t *testing.T) {
	fixture := newCommandFixture(t)
	ctx := context.Background()
	client, err := fixture.factory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	content := "\n== Today\nUna idea con acentos y espacios  \n\n"
	created, err := client.CreateJournal(ctx, journal.CreateCommand{Date: "2026-10-04", Content: content, Tags: []string{"Daily"}})
	if err != nil {
		t.Fatal(err)
	}
	id := created.Journal.ID()
	loaded, err := client.GetJournal(ctx, journal.GetQuery{ID: id.String()})
	if err != nil || loaded.Journal.Content() != content {
		t.Fatalf("created content was not preserved: %+v %v", loaded, err)
	}
	updatedContent := content + "Otra línea\n:overmind-id: body text, not metadata\n"
	updated, err := client.UpdateJournal(ctx, journal.UpdateCommand{ID: id.String(), Content: updatedContent, Tags: []string{"Work"}})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Journal.ID() != id || !updated.Journal.Date().Equal(created.Journal.Date()) || !updated.Journal.Metadata().CreatedAt().Equal(created.Journal.Metadata().CreatedAt()) || updated.Journal.Metadata().UpdatedAt().Before(created.Journal.Metadata().UpdatedAt()) {
		t.Fatal("update changed identity, date, or creation time")
	}
	loaded, err = client.GetJournal(ctx, journal.GetQuery{ID: id.String()})
	if err != nil || loaded.Journal.Content() != updatedContent || !reflect.DeepEqual(loaded.Journal.Tags().Strings(), []string{"work"}) {
		t.Fatalf("updated document was not reloaded: %+v %v", loaded, err)
	}
	path := filepath.Join(fixture.vault, "notes", id.String()+".adoc")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := (codecs.JournalCodec{}).Decode(source)
	if err != nil || decoded.Content() != updatedContent {
		t.Fatalf("document body was not preserved: %v", err)
	}
	if _, err := client.CreateJournal(ctx, journal.CreateCommand{Date: "2026-10-04", Content: "Overwrite"}); !errors.Is(err, journal.ErrAlreadyExists) {
		t.Fatalf("duplicate creation should remain rejected: %v", err)
	}
	if _, err := client.UpdateJournal(ctx, journal.UpdateCommand{ID: id.String(), Content: "Invalid update", Tags: []string{"!!!"}}); err == nil {
		t.Fatal("invalid tags were saved")
	}
	if _, err := client.UpdateJournal(ctx, journal.UpdateCommand{ID: uuid.New().String(), Content: "Missing"}); !errors.Is(err, ports.ErrNoteNotFound) {
		t.Fatalf("update created an absent note: %v", err)
	}
	current, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(current, source) {
		t.Fatal("rejected creation or update changed the saved document")
	}
	var storedDate string
	if err := fixture.db().QueryRow("SELECT date FROM journals WHERE note_id=?", id.String()).Scan(&storedDate); err != nil || storedDate != "2026-10-04" {
		t.Fatal("journal index did not preserve the date")
	}
	fixture.requireSingleNote()
}
