package person

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/ports"
)

// GetQuery identifies a person by its stable UUID.
type GetQuery struct {
	ID string
}

// GetResult contains the person returned after a successful get operation.
type GetResult struct {
	Person *persons.Person
}

// GetHandler loads a person from its source document and verifies the requested identity.
type GetHandler struct {
	repository ports.PersonRepository
}

// NewGetHandler creates the use-case handler with its dependencies.
func NewGetHandler(repository ports.PersonRepository) *GetHandler {
	return &GetHandler{
		repository: repository,
	}
}

// Handle rejects invalid UUIDs and mismatched repository responses before returning a person.
func (handler *GetHandler) Handle(ctx context.Context, query GetQuery) (GetResult, error) {
	if err := ctx.Err(); err != nil {
		return GetResult{}, err
	}

	id, err := uuid.Parse(query.ID)
	if err != nil || id == uuid.Nil() {
		return GetResult{}, fmt.Errorf("valid person ID is required")
	}

	if handler.repository == nil {
		return GetResult{}, fmt.Errorf("person repository is not configured")
	}

	entity, err := handler.repository.ByID(ctx, id)
	if err != nil {
		return GetResult{}, fmt.Errorf("get person: %w", err)
	}

	if entity == nil || entity.ID() != id {
		return GetResult{}, fmt.Errorf("person identity does not match requested ID")
	}

	return GetResult{Person: entity}, nil
}
