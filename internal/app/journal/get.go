package journal

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/ports"
)

// GetQuery identifies a journal by its stable UUID.
type GetQuery struct {
	ID string
}

// GetResult contains the journal returned after a successful get operation.
type GetResult struct {
	Journal *journals.Journal
}

// GetHandler loads a journal from its source document and verifies the requested identity.
type GetHandler struct {
	repository ports.JournalRepository
}

// NewGetHandler creates the use-case handler with its dependencies.
func NewGetHandler(repository ports.JournalRepository) *GetHandler {
	return &GetHandler{
		repository: repository,
	}
}

// Handle rejects invalid UUIDs and mismatched repository responses before returning a journal.
func (handler *GetHandler) Handle(ctx context.Context, query GetQuery) (GetResult, error) {
	if err := ctx.Err(); err != nil {
		return GetResult{}, err
	}

	id, err := uuid.Parse(query.ID)
	if err != nil || id == uuid.Nil() {
		return GetResult{}, fmt.Errorf("valid journal ID is required")
	}

	if handler.repository == nil {
		return GetResult{}, fmt.Errorf("journal repository is not configured")
	}

	entity, err := handler.repository.ByID(ctx, id)
	if err != nil {
		return GetResult{}, fmt.Errorf("get journal: %w", err)
	}

	if entity == nil || entity.ID() != id {
		return GetResult{}, fmt.Errorf("journal identity does not match requested ID")
	}

	return GetResult{Journal: entity}, nil
}
