package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/id"
)

// ErrJournalAlreadyExists indicates that a journal exists for a date.
var ErrJournalAlreadyExists = errors.New("journal already exists")

// CreateJournal creates, persists and indexes today's journal note.
func (app *App) CreateJournal(ctx context.Context, tags domain.Tags) (string, error) {
	now := time.Now()
	date := domain.DateFromTime(now)
	exists, err := app.index.JournalExists(ctx, date)
	if err != nil {
		return "", fmt.Errorf("check journal existence: %w", err)
	}
	if exists {
		return "", ErrJournalAlreadyExists
	}

	journal, err := domain.NewJournal(id.New(), date, tags, now)
	if err != nil {
		return "", err
	}

	return app.save(ctx, journal)
}
