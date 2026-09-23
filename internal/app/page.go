package app

import (
	"context"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/id"
)

// CreatePage creates, persists and indexes a page.
func (app *App) CreatePage(
	ctx context.Context,
	title domain.Title,
	area domain.Area,
	tags domain.Tags,
) (string, error) {
	page, err := domain.NewPage(id.New(), title, area, tags, time.Now())
	if err != nil {
		return "", err
	}

	return app.save(ctx, page)
}
