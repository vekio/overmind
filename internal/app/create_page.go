package app

import (
	"context"
	"fmt"
	"time"

	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// CreatePageCommand requires a title; area and tags are optional raw inputs.
type CreatePageCommand struct {
	Title string
	Area  string
	Tags  []string
}
type CreatePageResult struct{ Page *pages.Page }
type CreatePageHandler struct {
	repository ports.PageRepository
	ids        ports.IDGenerator
}

func newCreatePageHandler(repository ports.PageRepository, ids ports.IDGenerator) *CreatePageHandler {
	return &CreatePageHandler{repository: repository, ids: ids}
}
func (handler *CreatePageHandler) Handle(ctx context.Context, command CreatePageCommand) (CreatePageResult, error) {
	if err := ctx.Err(); err != nil {
		return CreatePageResult{}, err
	}
	if handler.repository == nil || handler.ids == nil {
		return CreatePageResult{}, fmt.Errorf("create page dependencies are not configured")
	}
	title, err := shared.NewTitle(command.Title)
	if err != nil {
		return CreatePageResult{}, err
	}
	var area pages.Area
	if command.Area != "" {
		area, err = pages.NewArea(command.Area)
		if err != nil {
			return CreatePageResult{}, err
		}
	}

	values := make([]shared.Tag, 0, len(command.Tags))
	for _, raw := range command.Tags {
		tag, err := shared.NewTag(raw)
		if err != nil {
			return CreatePageResult{}, err
		}
		values = append(values, tag)
	}
	tags, err := shared.NewTags(values...)
	if err != nil {
		return CreatePageResult{}, err
	}
	now := time.Now()
	metadata, err := shared.NewEntityMetadata(now, now)
	if err != nil {
		return CreatePageResult{}, err
	}
	entity, err := pages.NewPage(handler.ids.Generate(), title, area, tags, metadata)
	if err != nil {
		return CreatePageResult{}, err
	}
	if err := handler.repository.Save(ctx, entity); err != nil {
		return CreatePageResult{}, fmt.Errorf("save page: %w", err)
	}
	return CreatePageResult{Page: entity}, nil
}
