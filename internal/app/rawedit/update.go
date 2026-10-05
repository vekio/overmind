package rawedit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/ports"
)

var (
	// ErrConflict means the vault no longer matches the caller's source snapshot.
	ErrConflict = errors.New("note changed in the vault while editing")
	// ErrIndexUpdate means the source was saved but its projection needs a retry.
	ErrIndexUpdate = errors.New("document saved but index update failed")
)

// UpdateCommand submits complete source against the last known vault snapshot.
// Protected identity and creation metadata must remain unchanged.
type UpdateCommand struct {
	ID   string
	Kind ports.NoteKind

	// Original is the exact source read at load time or returned after a partial save.
	Original []byte
	Source   []byte

	// IndexOnly retries a failed projection of Original without writing the file.
	IndexOnly bool
}

// UpdateResult reports the authoritative source even when indexing fails.
// Callers must retain that snapshot before retrying an IndexPending result.
type UpdateResult struct {
	// Changed marks a persisted edit or an indexing retry, rather than a no-op.
	Changed bool

	// Note is available after a no-op, a source write or an index-only retry.
	Note ports.Note

	// IndexPending is true only when the file is saved but projection failed.
	IndexPending bool
}

// UpdateHandler validates and saves source bytes, then updates their projection.
// File persistence and index repair are separate stages with distinct outcomes.
type UpdateHandler struct {
	store ports.NoteStore
	codec ports.RawNoteCodec
	index ports.Index
}

// NewUpdateHandler creates the raw-source writer with its storage, codec and index.
func NewUpdateHandler(store ports.NoteStore, codec ports.RawNoteCodec, index ports.Index) *UpdateHandler {
	return &UpdateHandler{store: store, codec: codec, index: index}
}

// Handle checks the source snapshot and protected metadata before saving and projecting.
// IndexOnly retries projection without changing source bytes or the update timestamp.
func (handler *UpdateHandler) Handle(ctx context.Context, command UpdateCommand) (UpdateResult, error) {
	if err := ctx.Err(); err != nil {
		return UpdateResult{}, err
	}
	id, err := uuid.Parse(command.ID)
	if err != nil || id == uuid.Nil() {
		return UpdateResult{}, fmt.Errorf("valid note ID is required")
	}
	if handler.store == nil || handler.codec == nil || handler.index == nil {
		return UpdateResult{}, fmt.Errorf("raw editing is not configured")
	}
	current, err := handler.store.Get(ctx, id)
	if err != nil {
		return UpdateResult{}, err
	}
	if !bytes.Equal(current.Content, command.Original) {
		return UpdateResult{}, ErrConflict
	}
	original, err := handler.codec.Inspect(current.Content)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("read original note: %w", err)
	}
	if current.ID != id || original.ID != id || (command.Kind != "" && original.Kind != command.Kind) {
		return UpdateResult{}, fmt.Errorf("note identity or type does not match original")
	}
	if command.IndexOnly {
		if _, err := handler.codec.Decode(original.Kind, current.Content); err != nil {
			return UpdateResult{}, err
		}
		return handler.project(ctx, current)
	}
	if bytes.Equal(command.Source, current.Content) {
		return UpdateResult{Note: current}, nil
	}
	draft, err := handler.codec.Decode(original.Kind, command.Source)
	if err != nil {
		return UpdateResult{}, err
	}
	if draft.ID != id || draft.Kind != original.Kind {
		return UpdateResult{}, fmt.Errorf("note ID and type cannot change")
	}
	if !original.Metadata.CreatedAt().Equal(draft.Metadata.CreatedAt()) {
		return UpdateResult{}, fmt.Errorf("note creation time cannot change")
	}
	if draft.Date != original.Date {
		return UpdateResult{}, fmt.Errorf("journal date cannot change")
	}
	now := time.Now()
	// A clock adjustment must not move the managed update time backwards.
	if now.Before(original.Metadata.UpdatedAt()) {
		now = original.Metadata.UpdatedAt()
	}
	source, err := handler.codec.Stamp(command.Source, now)
	if err != nil {
		return UpdateResult{}, err
	}
	// Validate the exact bytes that will be persisted, including managed metadata.
	if _, err := handler.codec.Decode(original.Kind, source); err != nil {
		return UpdateResult{}, err
	}
	// Parsing can take time; check again immediately before the atomic file write.
	latest, err := handler.store.Get(ctx, id)
	if err != nil {
		return UpdateResult{}, err
	}
	if !bytes.Equal(latest.Content, current.Content) {
		return UpdateResult{}, ErrConflict
	}
	note := ports.Note{ID: id, Content: source}
	path, err := handler.store.Put(ctx, note)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("save raw note: %w", err)
	}
	note.Path = path
	return handler.project(ctx, note)
}

func (handler *UpdateHandler) project(ctx context.Context, note ports.Note) (UpdateResult, error) {
	result := UpdateResult{Changed: true, Note: note}
	if err := handler.index.Project(ctx, note); err != nil {
		result.IndexPending = true
		return result, fmt.Errorf("%w: %w; retry indexing", ErrIndexUpdate, err)
	}
	return result, nil
}
