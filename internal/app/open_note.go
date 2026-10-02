package app

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/ports"
)

// OpenNoteHandler reads an existing note without changing it.
type OpenNoteHandler struct {
	reader ports.NoteReader
}

func newOpenNoteHandler(reader ports.NoteReader) *OpenNoteHandler {
	return &OpenNoteHandler{reader: reader}
}

func (handler *OpenNoteHandler) Handle(ctx context.Context, id uuid.UUID) ([]byte, error) {
	if id == uuid.Nil() {
		return nil, fmt.Errorf("note ID is required")
	}
	source, err := handler.reader.Read(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("read note %s: %w", id, err)
	}
	return source, nil
}
