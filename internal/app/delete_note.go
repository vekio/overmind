package app

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/ports"
)

// DeleteNoteHandler removes a note document and its indexed metadata.
type DeleteNoteHandler struct {
	files ports.NoteDeleter
	index ports.NoteIndex
}

func newDeleteNoteHandler(files ports.NoteDeleter, index ports.NoteIndex) *DeleteNoteHandler {
	return &DeleteNoteHandler{files: files, index: index}
}

func (handler *DeleteNoteHandler) Handle(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil() {
		return fmt.Errorf("note ID is required")
	}
	if err := handler.files.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete note %s from vault: %w", id, err)
	}
	if err := handler.index.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete note %s from index: %w", id, err)
	}
	return nil
}
