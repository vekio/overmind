package app

import (
	"context"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

// CreatePageCommand contains the input for creating a page.
type CreatePageCommand struct {
	Title domain.Title
	Area  domain.Area
	Tags  domain.Tags
}

// CreatePageResult contains the created page and its path.
type CreatePageResult struct {
	Page domain.Page
	Path string
}

// CreatePageHandler creates page notes.
type CreatePageHandler struct {
	saver *noteSaver
	ids   ports.IDGenerator
}

func newCreatePageHandler(saver *noteSaver, ids ports.IDGenerator) *CreatePageHandler {
	return &CreatePageHandler{saver: saver, ids: ids}
}

// Handle creates, persists and indexes a page.
func (handler *CreatePageHandler) Handle(ctx context.Context, command CreatePageCommand) (CreatePageResult, error) {
	page, err := domain.NewPage(handler.ids.Generate(), command.Title, command.Area, command.Tags, time.Now())
	if err != nil {
		return CreatePageResult{}, err
	}
	path, err := handler.saver.save(ctx, page)
	if err != nil {
		return CreatePageResult{}, err
	}

	return CreatePageResult{Page: page, Path: path}, nil
}
