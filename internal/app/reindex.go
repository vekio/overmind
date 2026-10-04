package app

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/ports"
)

type ReindexCommand struct{}
type ReindexResult struct{ Indexed int }

type ReindexHandler struct {
	scanner   ports.NoteScanner
	projector ports.NoteProjector
	index     ports.IndexRebuilder
}

func newReindexHandler(scanner ports.NoteScanner, projector ports.NoteProjector, index ports.IndexRebuilder) *ReindexHandler {
	return &ReindexHandler{scanner: scanner, projector: projector, index: index}
}

// Handle reconstructs the complete index from documents. A failed scan or
// projection rolls back the replacement, including removals of stale entries.
func (handler *ReindexHandler) Handle(ctx context.Context, _ ReindexCommand) (ReindexResult, error) {
	if err := ctx.Err(); err != nil {
		return ReindexResult{}, err
	}
	if handler.scanner == nil || handler.projector == nil || handler.index == nil {
		return ReindexResult{}, fmt.Errorf("reindex dependencies are not configured")
	}
	count := 0
	seen := make(map[uuid.UUID]string)
	err := handler.index.Rebuild(ctx, func(index ports.Index) error {
		return handler.scanner.Scan(ctx, func(note ports.Note) error {
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
			if err := handler.projector.Project(ctx, note, index); err != nil {
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
