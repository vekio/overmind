package app

import (
	"context"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/id"
)

// CreateJournal creates, persists and indexes today's journal note.
func (app *App) CreateJournal(ctx context.Context, tags domain.Tags) (string, error) {
	now := time.Now()
	journal, err := domain.NewJournal(id.New(), domain.DateFromTime(now), tags, now)
	if err != nil {
		return "", err
	}

	return app.save(ctx, journal)
}
