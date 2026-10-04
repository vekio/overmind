package cli

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/domain/calendar"
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
	if _, err := fixture.run([]string{"journal", "--date", date, "--tag", "Different"}, ""); !errors.Is(err, app.ErrJournalAlreadyExists) {
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
