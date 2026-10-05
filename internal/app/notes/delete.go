package notes

import (
	"context"
	"fmt"
	"strings"
	"uuid"

	"github.com/vekio/overmind/internal/ports"
)

// DeleteCommand identifies any kind of note using raw input.
type DeleteCommand struct {
	ID string
}

// DeleteResult identifies the note whose source and projection have been removed.
type DeleteResult struct {
	ID uuid.UUID
}

// DeleteHandler removes a source document before deleting its searchable projection.
type DeleteHandler struct {
	store ports.NoteStore
	index ports.Index
}

// NewDeleteHandler creates the use-case handler with its dependencies.
func NewDeleteHandler(store ports.NoteStore, index ports.Index) *DeleteHandler {
	return &DeleteHandler{
		store: store,
		index: index,
	}
}

// Handle removes the source document before its derived index entry. Both
// removals are idempotent, so retrying also repairs a previous partial failure.
func (handler *DeleteHandler) Handle(ctx context.Context, command DeleteCommand) (DeleteResult, error) {
	if err := ctx.Err(); err != nil {
		return DeleteResult{}, err
	}
	id, err := uuid.Parse(strings.TrimSpace(command.ID))
	if err != nil {
		return DeleteResult{}, fmt.Errorf("invalid note ID: %w", err)
	}
	if id == uuid.Nil() {
		return DeleteResult{}, fmt.Errorf("note ID must not be nil")
	}
	if handler.store == nil || handler.index == nil {
		return DeleteResult{}, fmt.Errorf("delete note dependencies are not configured")
	}
	if err := handler.store.Delete(ctx, id); err != nil {
		return DeleteResult{}, fmt.Errorf("delete note document: %w", err)
	}
	if err := handler.index.Delete(ctx, id); err != nil {
		return DeleteResult{}, fmt.Errorf("note document removed but index update failed; retry delete or run overmind reindex: %w", err)
	}
	return DeleteResult{ID: id}, nil
}
