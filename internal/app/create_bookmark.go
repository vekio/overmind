package app

import (
	"context"
	"fmt"
	"time"

	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

type CreateBookmarkCommand struct {
	URL  string
	Tags []string
}
type CreateBookmarkResult struct{ Bookmark *bookmarks.Bookmark }
type CreateBookmarkHandler struct {
	repository ports.BookmarkRepository
	ids        ports.IDGenerator
}

func newCreateBookmarkHandler(repository ports.BookmarkRepository, ids ports.IDGenerator) *CreateBookmarkHandler {
	return &CreateBookmarkHandler{repository: repository, ids: ids}
}
func (handler *CreateBookmarkHandler) Handle(ctx context.Context, command CreateBookmarkCommand) (CreateBookmarkResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateBookmarkResult{}, err
	}
	if handler.repository == nil || handler.ids == nil {
		return CreateBookmarkResult{}, fmt.Errorf("create bookmark dependencies are not configured")
	}
	value, err := bookmarks.NewURL(command.URL)
	if err != nil {
		return CreateBookmarkResult{}, err
	}
	values := make([]shared.Tag, 0, len(command.Tags))
	for _, raw := range command.Tags {
		tag, err := shared.NewTag(raw)
		if err != nil {
			return CreateBookmarkResult{}, err
		}
		values = append(values, tag)
	}
	tags, err := shared.NewTags(values...)
	if err != nil {
		return CreateBookmarkResult{}, err
	}
	now := time.Now()
	metadata, err := shared.NewEntityMetadata(now, now)
	if err != nil {
		return CreateBookmarkResult{}, err
	}
	entity, err := bookmarks.NewBookmark(handler.ids.Generate(), value, tags, metadata)
	if err != nil {
		return CreateBookmarkResult{}, err
	}
	if err := handler.repository.Save(ctx, entity); err != nil {
		return CreateBookmarkResult{}, fmt.Errorf("save bookmark: %w", err)
	}
	return CreateBookmarkResult{Bookmark: entity}, nil
}
