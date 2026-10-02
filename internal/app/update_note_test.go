package app_test

import (
	"bytes"
	"context"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/app"
)

func TestUpdateNoteRejectsIdentityAndJournalDateChanges(t *testing.T) {
	ctx := context.Background()
	application := newAppFixture(t).application
	journal, err := application.Commands.CreateJournal.Handle(ctx, app.CreateJournalCommand{})
	if err != nil {
		t.Fatal(err)
	}
	id := journal.Journal.Metadata().ID()
	original, err := application.Queries.OpenNote.Handle(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	wrongID := bytes.Replace(original, []byte(id.String()), []byte(uuid.New().String()), 1)
	wrongDate := bytes.Replace(original, []byte(":overmind-date: "+journal.Journal.Date().String()), []byte(":overmind-date: 2000-01-01"), 1)
	for name, source := range map[string][]byte{"identity": wrongID, "journal date": wrongDate} {
		t.Run(name, func(t *testing.T) {
			if _, err := application.Commands.UpdateNote.Handle(ctx, app.UpdateNoteCommand{ID: id, Original: original, Source: source}); err == nil {
				t.Fatal("invalid edit was saved")
			}
			current, err := application.Queries.OpenNote.Handle(ctx, id)
			if err != nil || !bytes.Equal(current, original) {
				t.Fatalf("invalid edit changed source: %v", err)
			}
		})
	}
}
