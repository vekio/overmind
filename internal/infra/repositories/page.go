package repositories

import (
	"context"
	"fmt"

	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/ports"
)

var _ ports.PageRepository = (*PageRepository)(nil)

type PageRepository struct {
	notes ports.NoteStore
	index ports.Index
	codec ports.PageEncoder
}

func NewPageRepository(notes ports.NoteStore, index ports.Index, codec ports.PageEncoder) *PageRepository {
	return &PageRepository{notes: notes, index: index, codec: codec}
}
func (repository *PageRepository) Save(ctx context.Context, entity *pages.Page) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if repository.notes == nil || repository.index == nil || repository.codec == nil {
		return fmt.Errorf("page repository dependencies are not configured")
	}
	source, err := repository.codec.Encode(entity)
	if err != nil {
		return fmt.Errorf("encode page: %w", err)
	}
	path, err := repository.notes.Put(ctx, ports.Note{ID: entity.ID(), Content: source})
	if err != nil {
		return fmt.Errorf("store page: %w", err)
	}
	if err := repository.index.UpsertPage(ctx, entity, path); err != nil {
		return fmt.Errorf("update page index (document saved): %w", err)
	}
	return nil
}
