package inbox

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/ports"
)

// GetQuery identifies a inbox by its stable UUID.
type GetQuery struct {
	ID string
}

// GetResult contains the inbox returned after a successful get operation.
type GetResult struct {
	Inbox *inbox.Inbox
}

// GetHandler loads a inbox from its source document and verifies the requested identity.
type GetHandler struct {
	repository ports.InboxRepository
}

// NewGetHandler creates the use-case handler with its dependencies.
func NewGetHandler(repository ports.InboxRepository) *GetHandler {
	return &GetHandler{
		repository: repository,
	}
}

// Handle rejects invalid UUIDs and mismatched repository responses before returning a inbox.
func (handler *GetHandler) Handle(ctx context.Context, query GetQuery) (GetResult, error) {
	if err := ctx.Err(); err != nil {
		return GetResult{}, err
	}

	id, err := uuid.Parse(query.ID)
	if err != nil || id == uuid.Nil() {
		return GetResult{}, fmt.Errorf("valid inbox ID is required")
	}

	if handler.repository == nil {
		return GetResult{}, fmt.Errorf("inbox repository is not configured")
	}

	entity, err := handler.repository.ByID(ctx, id)
	if err != nil {
		return GetResult{}, fmt.Errorf("get inbox: %w", err)
	}

	if entity == nil || entity.ID() != id {
		return GetResult{}, fmt.Errorf("inbox identity does not match requested ID")
	}

	return GetResult{Inbox: entity}, nil
}
