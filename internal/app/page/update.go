package page

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// UpdateCommand requires a title; content, area and tags are optional raw inputs.
type UpdateCommand struct {
	ID      string
	Title   string
	Content string
	Area    string
	Tags    []string
}

// UpdateResult contains the page returned after a successful update operation.
type UpdateResult struct {
	Page *pages.Page
}

// UpdateHandler validates replacement values and retains the page's identity and creation time.
type UpdateHandler struct {
	repository ports.PageRepository
}

// NewUpdateHandler creates the use-case handler with its dependencies.
func NewUpdateHandler(repository ports.PageRepository) *UpdateHandler {
	return &UpdateHandler{
		repository: repository,
	}
}

// Handle loads the existing page, validates its replacement and persists the updated entity.
func (handler *UpdateHandler) Handle(ctx context.Context, command UpdateCommand) (UpdateResult, error) {
	if err := ctx.Err(); err != nil {
		return UpdateResult{}, err
	}
	if handler.repository == nil {
		return UpdateResult{}, fmt.Errorf("update page dependencies are not configured")
	}
	title, err := shared.NewTitle(command.Title)
	if err != nil {
		return UpdateResult{}, err
	}
	var area pages.Area
	if command.Area != "" {
		area, err = pages.NewArea(command.Area)
		if err != nil {
			return UpdateResult{}, err
		}
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
		return UpdateResult{}, fmt.Errorf("valid page ID is required")
	}
	existing, err := handler.repository.ByID(ctx, id)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("read page: %w", err)
	}
	if existing == nil || existing.ID() != id {
		return UpdateResult{}, fmt.Errorf("page identity does not match requested ID")
	}
	now := time.Now()
	metadata, err := existing.Metadata().Updated(now)
	if err != nil {
		return UpdateResult{}, err
	}
	entity, err := pages.NewPage(id, title, area, tags, metadata)
	if err != nil {
		return UpdateResult{}, err
	}
	if err := entity.Rewrite(command.Content, now); err != nil {
		return UpdateResult{}, err
	}
	if err := handler.repository.Update(ctx, entity); err != nil {
		return UpdateResult{}, fmt.Errorf("update page: %w", err)
	}
	return UpdateResult{Page: entity}, nil
}
