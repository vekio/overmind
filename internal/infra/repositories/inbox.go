package repositories

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/ports"
)

var _ ports.InboxRepository = (*InboxRepository)(nil)

// InboxRepository stores managed inbox documents and their derived index entries.
type InboxRepository struct {
	notes ports.NoteStore
	index ports.Index
	codec ports.InboxCodec
}

// NewInboxRepository binds source storage, index and typed document codec.
func NewInboxRepository(notes ports.NoteStore, index ports.Index, codec ports.InboxCodec) *InboxRepository {
	return &InboxRepository{notes: notes, index: index, codec: codec}
}

// Save writes the inbox document before updating its index projection.
// An indexing error can occur after the document has been saved.
func (repository *InboxRepository) Save(ctx context.Context, entity *inbox.Inbox) error {
	return repository.persist(ctx, entity)
}

func (repository *InboxRepository) persist(ctx context.Context, entity *inbox.Inbox) error {
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

// ByID decodes the authoritative source and verifies its inbox kind and UUID.
func (repository *InboxRepository) ByID(ctx context.Context, id uuid.UUID) (*inbox.Inbox, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id == uuid.Nil() || repository.notes == nil || repository.codec == nil {
		return nil, fmt.Errorf("inbox identity and repository dependencies are required")
	}
	note, err := repository.notes.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("read inbox: %w", err)
	}
	entity, err := repository.codec.Decode(note.Content)
	if err != nil {
		return nil, fmt.Errorf("decode inbox: %w", err)
	}
	if entity == nil || note.ID != id || entity.ID() != id {
		return nil, fmt.Errorf("inbox identity does not match requested ID")
	}
	return entity, nil
}

// Update requires an existing inbox and preserves its creation time.
// It re-encodes the document before refreshing the projection.
func (repository *InboxRepository) Update(ctx context.Context, entity *inbox.Inbox) error {
	if entity == nil || repository.index == nil {
		return fmt.Errorf("inbox and repository dependencies are required")
	}
	existing, err := repository.ByID(ctx, entity.ID())
	if err != nil {
		return err
	}
	if !existing.Metadata().CreatedAt().Equal(entity.Metadata().CreatedAt()) {
		return fmt.Errorf("inbox creation time cannot change")
	}
	return repository.persist(ctx, entity)
}
