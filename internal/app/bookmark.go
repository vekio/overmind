package app

import (
	"context"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/id"
)

// CreateBookmark creates, persists and indexes a bookmark.
func (app *App) CreateBookmark(
	ctx context.Context,
	url domain.URL,
	tags domain.Tags,
) (string, error) {
	bookmark, err := domain.NewBookmark(id.New(), url, tags, time.Now())
	if err != nil {
		return "", err
	}

	return app.save(ctx, bookmark)
}
