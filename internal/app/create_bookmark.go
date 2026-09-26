package app

import (
	"context"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

// CreateBookmarkCommand contains the input for creating a bookmark.
type CreateBookmarkCommand struct {
	URL  domain.URL
	Tags domain.Tags
}

// CreateBookmarkResult contains the created bookmark and its path.
type CreateBookmarkResult struct {
	Bookmark domain.Bookmark
	Path     string
}

// CreateBookmarkHandler creates bookmark notes.
type CreateBookmarkHandler struct {
	saver *noteSaver
	ids   ports.IDGenerator
}

func newCreateBookmarkHandler(saver *noteSaver, ids ports.IDGenerator) *CreateBookmarkHandler {
	return &CreateBookmarkHandler{saver: saver, ids: ids}
}

// Handle creates, persists and indexes a bookmark.
func (handler *CreateBookmarkHandler) Handle(ctx context.Context, command CreateBookmarkCommand) (CreateBookmarkResult, error) {
	bookmark, err := domain.NewBookmark(handler.ids.Generate(), command.URL, command.Tags, time.Now())
	if err != nil {
		return CreateBookmarkResult{}, err
	}
	path, err := handler.saver.save(ctx, bookmark)
	if err != nil {
		return CreateBookmarkResult{}, err
	}

	return CreateBookmarkResult{Bookmark: bookmark, Path: path}, nil
}
