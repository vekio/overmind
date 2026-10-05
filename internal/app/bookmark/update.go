package bookmark

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// UpdateCommand replaces the editable fields of an existing bookmark.
// Empty optional values clear their previous contents.
type UpdateCommand struct {
	Content string
	ID      string
	URL     string
	Tags    []string
}

// UpdateResult contains the bookmark returned after a successful update operation.
type UpdateResult struct {
	Bookmark *bookmarks.Bookmark
}

// UpdateHandler validates replacement values and retains the bookmark's identity and creation time.
type UpdateHandler struct {
	repository ports.BookmarkRepository
}

// NewUpdateHandler creates the use-case handler with its dependencies.
func NewUpdateHandler(repository ports.BookmarkRepository) *UpdateHandler {
	return &UpdateHandler{
		repository: repository,
	}
}

// Handle loads the existing bookmark, validates its replacement and persists the updated entity.
func (handler *UpdateHandler) Handle(ctx context.Context, command UpdateCommand) (UpdateResult, error) {
	if err := ctx.Err(); err != nil {
		return UpdateResult{}, err
	}
	if handler.repository == nil {
		return UpdateResult{}, fmt.Errorf("update bookmark dependencies are not configured")
	}
	value, err := bookmarks.NewURL(command.URL)
	if err != nil {
		return UpdateResult{}, err
	}
	values := make([]shared.Tag, 0, len(command.Tags))
	for _, raw := range command.Tags {
		tag, err := shared.NewTag(raw)
		if err != nil {
			return UpdateResult{}, err
		}
		values = append(values, tag)
	}
	tags, err := shared.NewTags(values...)
	if err != nil {
		return UpdateResult{}, err
	}
	id, err := uuid.Parse(command.ID)
	if err != nil || id == uuid.Nil() {
		return UpdateResult{}, fmt.Errorf("valid bookmark ID is required")
	}
	existing, err := handler.repository.ByID(ctx, id)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("read bookmark: %w", err)
	}
	if existing == nil || existing.ID() != id {
		return UpdateResult{}, fmt.Errorf("bookmark identity does not match requested ID")
	}
	now := time.Now()
	metadata, err := existing.Metadata().Updated(now)
	if err != nil {
		return UpdateResult{}, err
	}
	entity, err := bookmarks.NewBookmark(id, value, tags, metadata)
	if err != nil {
		return UpdateResult{}, err
	}
	if err := entity.Rewrite(command.Content, now); err != nil {
		return UpdateResult{}, err
	}
	if err := handler.repository.Update(ctx, entity); err != nil {
		return UpdateResult{}, fmt.Errorf("update bookmark: %w", err)
	}
	return UpdateResult{Bookmark: entity}, nil
}
