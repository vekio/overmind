package notes

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/ports"
)

// ReindexCommand requests a complete rebuild from the current vault documents.
type ReindexCommand struct{}

// ReindexResult counts documents indexed after a successful transaction.
type ReindexResult struct {
	Indexed int
}

// ReindexHandler scans the vault and rebuilds all projections in one index transaction.
type ReindexHandler struct {
	store ports.NoteStore
	index ports.Index
}

// NewReindexHandler creates the use-case handler with its dependencies.
func NewReindexHandler(store ports.NoteStore, index ports.Index) *ReindexHandler {
	return &ReindexHandler{
		store: store,
		index: index,
	}
}

// Handle reconstructs the complete index from documents. A failed scan or
// projection rolls back the replacement, including removals of stale entries.
func (handler *ReindexHandler) Handle(ctx context.Context, _ ReindexCommand) (ReindexResult, error) {
	if err := ctx.Err(); err != nil {
		return ReindexResult{}, err
	}
	if handler.store == nil || handler.index == nil {
		return ReindexResult{}, fmt.Errorf("reindex dependencies are not configured")
	}
	count := 0
	seen := make(map[uuid.UUID]string)
	err := handler.index.Rebuild(ctx, func(index ports.Index) error {
		return handler.store.Scan(ctx, func(note ports.Note) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if note.ID == uuid.Nil() {
				return fmt.Errorf("%s: note ID is required", note.Path)
			}
			if previous, ok := seen[note.ID]; ok {
				return fmt.Errorf("duplicate note ID %s in %s and %s", note.ID, previous, note.Path)
			}
			seen[note.ID] = note.Path
			if err := index.Project(ctx, note); err != nil {
				return fmt.Errorf("%s: %w", note.Path, err)
			}
			count++
			return nil
		})
	})
	if err != nil {
		return ReindexResult{}, fmt.Errorf("reindex failed; previous index preserved: %w", err)
	}
	return ReindexResult{Indexed: count}, nil
}
