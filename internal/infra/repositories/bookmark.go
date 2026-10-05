package repositories

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/ports"
)

var _ ports.BookmarkRepository = (*BookmarkRepository)(nil)

// BookmarkRepository stores managed bookmark documents and their derived index entries.
type BookmarkRepository struct {
	notes ports.NoteStore
	index ports.Index
	codec ports.BookmarkCodec
}

// NewBookmarkRepository binds source storage, index and typed document codec.
func NewBookmarkRepository(notes ports.NoteStore, index ports.Index, codec ports.BookmarkCodec) *BookmarkRepository {
	return &BookmarkRepository{notes: notes, index: index, codec: codec}
}

// Save writes the bookmark document before updating its index projection.
// An indexing error can occur after the document has been saved.
func (repository *BookmarkRepository) Save(ctx context.Context, entity *bookmarks.Bookmark) error {
	return repository.persist(ctx, entity)
}

func (repository *BookmarkRepository) persist(ctx context.Context, entity *bookmarks.Bookmark) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if repository.notes == nil || repository.index == nil || repository.codec == nil {
		return fmt.Errorf("bookmark repository dependencies are not configured")
	}
	source, err := repository.codec.Encode(entity)
	if err != nil {
		return fmt.Errorf("encode bookmark: %w", err)
	}
	path, err := repository.notes.Put(ctx, ports.Note{ID: entity.ID(), Content: source})
	if err != nil {
		return fmt.Errorf("store bookmark: %w", err)
	}
	if err := repository.index.UpsertBookmark(ctx, entity, path); err != nil {
		return fmt.Errorf("update bookmark index (document saved): %w", err)
	}
	return nil
}

// ByID decodes the authoritative source and verifies its bookmark kind and UUID.
func (repository *BookmarkRepository) ByID(ctx context.Context, id uuid.UUID) (*bookmarks.Bookmark, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id == uuid.Nil() || repository.notes == nil || repository.codec == nil {
		return nil, fmt.Errorf("bookmark identity and repository dependencies are required")
	}
	note, err := repository.notes.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("read bookmark: %w", err)
	}
	entity, err := repository.codec.Decode(note.Content)
	if err != nil {
		return nil, fmt.Errorf("decode bookmark: %w", err)
	}
	if entity == nil || note.ID != id || entity.ID() != id {
		return nil, fmt.Errorf("bookmark identity does not match requested ID")
	}
	return entity, nil
}

// Update requires an existing bookmark and preserves its creation time.
// It re-encodes the document before refreshing the projection.
func (repository *BookmarkRepository) Update(ctx context.Context, entity *bookmarks.Bookmark) error {
	if entity == nil || repository.index == nil {
		return fmt.Errorf("bookmark and repository dependencies are required")
	}
	existing, err := repository.ByID(ctx, entity.ID())
	if err != nil {
		return err
	}
	if !existing.Metadata().CreatedAt().Equal(entity.Metadata().CreatedAt()) {
		return fmt.Errorf("bookmark creation time cannot change")
	}
	return repository.persist(ctx, entity)
}
