package page

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/ports"
)

// GetQuery identifies a page by its stable UUID.
type GetQuery struct {
	ID string
}

// GetResult contains the page returned after a successful get operation.
type GetResult struct {
	Page *pages.Page
}

// GetHandler loads a page from its source document and verifies the requested identity.
type GetHandler struct {
	repository ports.PageRepository
}

// NewGetHandler creates the use-case handler with its dependencies.
func NewGetHandler(repository ports.PageRepository) *GetHandler {
	return &GetHandler{
		repository: repository,
	}
}

// Handle rejects invalid UUIDs and mismatched repository responses before returning a page.
func (handler *GetHandler) Handle(ctx context.Context, query GetQuery) (GetResult, error) {
	if err := ctx.Err(); err != nil {
		return GetResult{}, err
	}

	id, err := uuid.Parse(query.ID)
	if err != nil || id == uuid.Nil() {
		return GetResult{}, fmt.Errorf("valid page ID is required")
	}

	if handler.repository == nil {
		return GetResult{}, fmt.Errorf("page repository is not configured")
	}

	entity, err := handler.repository.ByID(ctx, id)
	if err != nil {
		return GetResult{}, fmt.Errorf("get page: %w", err)
	}

	if entity == nil || entity.ID() != id {
		return GetResult{}, fmt.Errorf("page identity does not match requested ID")
	}

	return GetResult{Page: entity}, nil
}
