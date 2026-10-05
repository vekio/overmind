package repositories

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/ports"
)

var _ ports.PageRepository = (*PageRepository)(nil)

// PageRepository stores managed page documents and their derived index entries.
type PageRepository struct {
	notes ports.NoteStore
	index ports.Index
	codec ports.PageCodec
}

// NewPageRepository binds source storage, index and typed document codec.
func NewPageRepository(notes ports.NoteStore, index ports.Index, codec ports.PageCodec) *PageRepository {
	return &PageRepository{notes: notes, index: index, codec: codec}
}

// Save writes the page document before updating its index projection.
// An indexing error can occur after the document has been saved.
func (repository *PageRepository) Save(ctx context.Context, entity *pages.Page) error {
	return repository.persist(ctx, entity)
}

func (repository *PageRepository) persist(ctx context.Context, entity *pages.Page) error {
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

// ByID decodes the authoritative source and verifies its page kind and UUID.
func (repository *PageRepository) ByID(ctx context.Context, id uuid.UUID) (*pages.Page, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id == uuid.Nil() || repository.notes == nil || repository.codec == nil {
		return nil, fmt.Errorf("page identity and repository dependencies are required")
	}
	note, err := repository.notes.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("read page: %w", err)
	}
	entity, err := repository.codec.Decode(note.Content)
	if err != nil {
		return nil, fmt.Errorf("decode page: %w", err)
	}
	if entity == nil || note.ID != id || entity.ID() != id {
		return nil, fmt.Errorf("page identity does not match requested ID")
	}
	return entity, nil
}

// Update requires an existing page and preserves its creation time.
// It re-encodes the document before refreshing the projection.
func (repository *PageRepository) Update(ctx context.Context, entity *pages.Page) error {
	if entity == nil || repository.index == nil {
		return fmt.Errorf("page and repository dependencies are required")
	}
	existing, err := repository.ByID(ctx, entity.ID())
	if err != nil {
		return err
	}
	if !existing.Metadata().CreatedAt().Equal(entity.Metadata().CreatedAt()) {
		return fmt.Errorf("page creation time cannot change")
	}
	return repository.persist(ctx, entity)
}
