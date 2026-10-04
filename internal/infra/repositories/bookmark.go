package repositories

import (
	"context"
	"fmt"

	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/ports"
)

var _ ports.BookmarkRepository = (*BookmarkRepository)(nil)

type BookmarkRepository struct {
	notes ports.NoteStore
	index ports.Index
	codec ports.BookmarkEncoder
}

func NewBookmarkRepository(notes ports.NoteStore, index ports.Index, codec ports.BookmarkEncoder) *BookmarkRepository {
	return &BookmarkRepository{notes: notes, index: index, codec: codec}
}
func (repository *BookmarkRepository) Save(ctx context.Context, entity *bookmarks.Bookmark) error {
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
