package bookmark

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/ports"
)

// GetQuery identifies a bookmark by its stable UUID.
type GetQuery struct {
	ID string
}

// GetResult contains the bookmark returned after a successful get operation.
type GetResult struct {
	Bookmark *bookmarks.Bookmark
}

// GetHandler loads a bookmark from its source document and verifies the requested identity.
type GetHandler struct {
	repository ports.BookmarkRepository
}

// NewGetHandler creates the use-case handler with its dependencies.
func NewGetHandler(repository ports.BookmarkRepository) *GetHandler {
	return &GetHandler{
		repository: repository,
	}
}

// Handle rejects invalid UUIDs and mismatched repository responses before returning a bookmark.
func (handler *GetHandler) Handle(ctx context.Context, query GetQuery) (GetResult, error) {
	if err := ctx.Err(); err != nil {
		return GetResult{}, err
	}

	id, err := uuid.Parse(query.ID)
	if err != nil || id == uuid.Nil() {
		return GetResult{}, fmt.Errorf("valid bookmark ID is required")
	}

	if handler.repository == nil {
		return GetResult{}, fmt.Errorf("bookmark repository is not configured")
	}

	entity, err := handler.repository.ByID(ctx, id)
	if err != nil {
		return GetResult{}, fmt.Errorf("get bookmark: %w", err)
	}

	if entity == nil || entity.ID() != id {
		return GetResult{}, fmt.Errorf("bookmark identity does not match requested ID")
	}

	return GetResult{Bookmark: entity}, nil
}
