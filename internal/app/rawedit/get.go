package rawedit

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/ports"
)

// GetQuery identifies the source document to open in an external editor.
type GetQuery struct {
	ID string
}

// GetResult includes the exact source and the kind declared in its header.
type GetResult struct {
	Note ports.Note
	Kind ports.NoteKind
}

// GetHandler checks identity and header metadata without requiring a valid body.
type GetHandler struct {
	store   ports.NoteStore
	decoder ports.RawNoteCodec
}

// NewGetHandler creates the raw-source reader with its storage and codec.
func NewGetHandler(store ports.NoteStore, decoder ports.RawNoteCodec) *GetHandler {
	return &GetHandler{store: store, decoder: decoder}
}

// Handle loads source and inspects its identity without requiring a valid document body.
func (handler *GetHandler) Handle(ctx context.Context, query GetQuery) (GetResult, error) {
	if err := ctx.Err(); err != nil {
		return GetResult{}, err
	}
	id, err := uuid.Parse(query.ID)
	if err != nil || id == uuid.Nil() {
		return GetResult{}, fmt.Errorf("valid note ID is required")
	}
	if handler.store == nil || handler.decoder == nil {
		return GetResult{}, fmt.Errorf("note storage is not configured")
	}
	note, err := handler.store.Get(ctx, id)
	if err != nil {
		return GetResult{}, err
	}
	draft, err := handler.decoder.Inspect(note.Content)
	if err != nil {
		return GetResult{}, err
	}
	if note.ID != id || draft.ID != id {
		return GetResult{}, fmt.Errorf("note identity does not match requested ID")
	}
	return GetResult{Note: note, Kind: draft.Kind}, nil
}
