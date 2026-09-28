package app_test

import (
	"context"
	"errors"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

type existingJournalIndex struct{ ports.NoteIndex }

func (existingJournalIndex) JournalExists(context.Context, domain.Date) (bool, error) {
	return true, nil
}

func TestCreateJournalRejectsExistingDate(t *testing.T) {
	application := app.New(app.Dependencies{Index: existingJournalIndex{}})
	if _, err := application.Commands.CreateJournal.Handle(context.Background(), app.CreateJournalCommand{}); !errors.Is(err, app.ErrJournalAlreadyExists) {
		t.Fatalf("existing journal = %v, want ErrJournalAlreadyExists", err)
	}
}
