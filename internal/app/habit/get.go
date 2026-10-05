package habit

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/ports"
)

// GetQuery identifies a habit by its stable UUID.
type GetQuery struct {
	ID string
}

// GetResult contains the habit returned after a successful get operation.
type GetResult struct {
	Habit *habits.Habit
}

// GetHandler loads a habit from its source document and verifies the requested identity.
type GetHandler struct {
	repository ports.HabitRepository
}

// NewGetHandler creates the use-case handler with its dependencies.
func NewGetHandler(repository ports.HabitRepository) *GetHandler {
	return &GetHandler{
		repository: repository,
	}
}

// Handle rejects invalid UUIDs and mismatched repository responses before returning a habit.
func (handler *GetHandler) Handle(ctx context.Context, query GetQuery) (GetResult, error) {
	if err := ctx.Err(); err != nil {
		return GetResult{}, err
	}

	id, err := uuid.Parse(query.ID)
	if err != nil || id == uuid.Nil() {
		return GetResult{}, fmt.Errorf("valid habit ID is required")
	}

	if handler.repository == nil {
		return GetResult{}, fmt.Errorf("habit repository is not configured")
	}

	entity, err := handler.repository.ByID(ctx, id)
	if err != nil {
		return GetResult{}, fmt.Errorf("get habit: %w", err)
	}

	if entity == nil || entity.ID() != id {
		return GetResult{}, fmt.Errorf("habit identity does not match requested ID")
	}

	return GetResult{Habit: entity}, nil
}
