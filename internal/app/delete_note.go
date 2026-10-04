package app

import (
	"context"
	"fmt"
	"strings"
	"uuid"

	"github.com/vekio/overmind/internal/ports"
)

// DeleteNoteCommand identifies any kind of note using raw input.
type DeleteNoteCommand struct{ ID string }
type DeleteNoteResult struct{ ID uuid.UUID }

type DeleteNoteHandler struct {
	store ports.NoteStore
	index ports.NoteIndexDeleter
}

func newDeleteNoteHandler(store ports.NoteStore, index ports.NoteIndexDeleter) *DeleteNoteHandler {
	return &DeleteNoteHandler{store: store, index: index}
}

// Handle removes the source document before its derived index entry. Both
// removals are idempotent, so retrying also repairs a previous partial failure.
func (handler *DeleteNoteHandler) Handle(ctx context.Context, command DeleteNoteCommand) (DeleteNoteResult, error) {
	if err := ctx.Err(); err != nil {
		return DeleteNoteResult{}, err
	}
	id, err := uuid.Parse(strings.TrimSpace(command.ID))
	if err != nil {
		return DeleteNoteResult{}, fmt.Errorf("invalid note ID: %w", err)
	}
	if id == uuid.Nil() {
		return DeleteNoteResult{}, fmt.Errorf("note ID must not be nil")
	}
	if handler.store == nil || handler.index == nil {
		return DeleteNoteResult{}, fmt.Errorf("delete note dependencies are not configured")
	}
	if err := handler.store.Delete(ctx, id); err != nil {
		return DeleteNoteResult{}, fmt.Errorf("delete note document: %w", err)
	}
	if err := handler.index.Delete(ctx, id); err != nil {
		return DeleteNoteResult{}, fmt.Errorf("note document removed but index update failed; retry delete or run overmind reindex: %w", err)
	}
	return DeleteNoteResult{ID: id}, nil
}
