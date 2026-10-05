package page

import (
	"context"
	"fmt"
	"time"

	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// CreateCommand requires a title; content, area and tags are optional raw inputs.
type CreateCommand struct {
	Title   string
	Content string
	Area    string
	Tags    []string
}

// CreateResult contains the page returned after a successful create operation.
type CreateResult struct {
	Page *pages.Page
}

// CreateHandler validates input, assigns identity and lifecycle metadata, and saves a page.
type CreateHandler struct {
	repository ports.PageRepository
	ids        ports.IDGenerator
}

// NewCreateHandler creates the use-case handler with its dependencies.
func NewCreateHandler(repository ports.PageRepository, ids ports.IDGenerator) *CreateHandler {
	return &CreateHandler{
		repository: repository,
		ids:        ids,
	}
}

// Handle validates raw input before persisting a new page and its index projection.
func (handler *CreateHandler) Handle(ctx context.Context, command CreateCommand) (CreateResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateResult{}, err
	}
	if handler.repository == nil || handler.ids == nil {
		return CreateResult{}, fmt.Errorf("create page dependencies are not configured")
	}
	title, err := shared.NewTitle(command.Title)
	if err != nil {
		return CreateResult{}, err
	}
	var area pages.Area
	if command.Area != "" {
		area, err = pages.NewArea(command.Area)
		if err != nil {
			return CreateResult{}, err
		}
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
	entity, err := pages.NewPage(handler.ids.Generate(), title, area, tags, metadata)
	if err != nil {
		return CreateResult{}, err
	}
	if err := entity.Rewrite(command.Content, now); err != nil {
		return CreateResult{}, err
	}
	if err := handler.repository.Save(ctx, entity); err != nil {
		return CreateResult{}, fmt.Errorf("save page: %w", err)
	}
	return CreateResult{Page: entity}, nil
}
