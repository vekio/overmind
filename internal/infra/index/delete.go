package index

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/ports"
)

var _ ports.NoteIndexDeleter = (*Index)(nil)

// Delete relies on foreign-key cascades to remove specialized data, note tags
// and person groups while preserving shared tag and group definitions.
func (index *Index) Delete(ctx context.Context, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id == uuid.Nil() {
		return fmt.Errorf("note ID is required")
	}
	if err := index.queries.DeleteNote(ctx, id.String()); err != nil {
		return fmt.Errorf("delete indexed note: %w", err)
	}
	return nil
}
