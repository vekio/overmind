package repositories

import (
	"context"
	"fmt"

	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/ports"
)

var _ ports.InboxRepository = (*InboxRepository)(nil)

type InboxRepository struct {
	notes ports.NoteStore
	index ports.Index
	codec ports.InboxEncoder
}

func NewInboxRepository(notes ports.NoteStore, index ports.Index, codec ports.InboxEncoder) *InboxRepository {
	return &InboxRepository{notes: notes, index: index, codec: codec}
}
func (repository *InboxRepository) Save(ctx context.Context, entity *inbox.Inbox) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if repository.notes == nil || repository.index == nil || repository.codec == nil {
		return fmt.Errorf("inbox repository dependencies are not configured")
	}
	source, err := repository.codec.Encode(entity)
	if err != nil {
		return fmt.Errorf("encode inbox: %w", err)
	}
	path, err := repository.notes.Put(ctx, ports.Note{ID: entity.ID(), Content: source})
	if err != nil {
		return fmt.Errorf("store inbox: %w", err)
	}
	if err := repository.index.UpsertInbox(ctx, entity, path); err != nil {
		return fmt.Errorf("update inbox index (document saved): %w", err)
	}
	return nil
}
