package app

import (
	"context"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/id"
)

// Capture creates, persists and indexes an inbox note.
func (app *App) Capture(ctx context.Context, content string) (string, error) {
	inbox, err := domain.NewInbox(id.New(), content, time.Now())
	if err != nil {
		return "", err
	}

	return app.save(ctx, inbox)
}
