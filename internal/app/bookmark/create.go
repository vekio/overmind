package bookmark

import (
	"context"
	"fmt"
	"time"

	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// CreateCommand supplies raw values for a new bookmark; optional collections may be empty.
type CreateCommand struct {
	Content string
	URL     string
	Tags    []string
}

// CreateResult contains the bookmark returned after a successful create operation.
type CreateResult struct {
	Bookmark *bookmarks.Bookmark
}

// CreateHandler validates input, assigns identity and lifecycle metadata, and saves a bookmark.
type CreateHandler struct {
	repository ports.BookmarkRepository
	ids        ports.IDGenerator
}

// NewCreateHandler creates the use-case handler with its dependencies.
func NewCreateHandler(repository ports.BookmarkRepository, ids ports.IDGenerator) *CreateHandler {
	return &CreateHandler{
		repository: repository,
		ids:        ids,
	}
}

// Handle validates raw input before persisting a new bookmark and its index projection.
func (handler *CreateHandler) Handle(ctx context.Context, command CreateCommand) (CreateResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateResult{}, err
	}
	if handler.repository == nil || handler.ids == nil {
		return CreateResult{}, fmt.Errorf("create bookmark dependencies are not configured")
	}
	value, err := bookmarks.NewURL(command.URL)
	if err != nil {
		return CreateResult{}, err
	}
	values := make([]shared.Tag, 0, len(command.Tags))
	for _, raw := range command.Tags {
		tag, err := shared.NewTag(raw)
		if err != nil {
			return CreateResult{}, err
		}
		values = append(values, tag)
	}
	tags, err := shared.NewTags(values...)
	if err != nil {
		return CreateResult{}, err
	}
	now := time.Now()
	metadata, err := shared.NewEntityMetadata(now, now)
	if err != nil {
		return CreateResult{}, err
	}
	entity, err := bookmarks.NewBookmark(handler.ids.Generate(), value, tags, metadata)
	if err != nil {
		return CreateResult{}, err
	}
	if err := entity.Rewrite(command.Content, now); err != nil {
		return CreateResult{}, err
	}
	if err := handler.repository.Save(ctx, entity); err != nil {
		return CreateResult{}, fmt.Errorf("save bookmark: %w", err)
	}
	return CreateResult{Bookmark: entity}, nil
}
