package repositories

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/ports"
)

var _ ports.JournalRepository = (*JournalRepository)(nil)

// JournalRepository stores managed journal documents and their derived index entries.
type JournalRepository struct {
	notes ports.NoteStore
	index ports.Index
	codec ports.JournalCodec
}

// NewJournalRepository binds source storage, index and typed document codec.
func NewJournalRepository(notes ports.NoteStore, index ports.Index, codec ports.JournalCodec) *JournalRepository {
	return &JournalRepository{notes: notes, index: index, codec: codec}
}

// Save rejects a date already present in the index, then writes the journal and its projection.
// An indexing error can occur after the document has been saved.
func (repository *JournalRepository) Save(ctx context.Context, entity *journals.Journal) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if repository.notes == nil || repository.index == nil || repository.codec == nil {
		return fmt.Errorf("journal repository dependencies are not configured")
	}
	if entity == nil {
		return fmt.Errorf("initialized journal is required")
	}
	exists, err := repository.index.JournalExists(ctx, entity.Date())
	if err != nil {
		return fmt.Errorf("check journal date: %w", err)
	}
	if exists {
		return ports.ErrJournalAlreadyExists
	}
	return repository.persist(ctx, entity)
}

// ByID decodes the authoritative source and verifies its journal kind and UUID.
func (repository *JournalRepository) ByID(ctx context.Context, id uuid.UUID) (*journals.Journal, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id == uuid.Nil() || repository.notes == nil || repository.codec == nil {
		return nil, fmt.Errorf("journal identity and repository dependencies are required")
	}
	note, err := repository.notes.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("read journal: %w", err)
	}
	entity, err := repository.codec.Decode(note.Content)
	if err != nil {
		return nil, fmt.Errorf("decode journal: %w", err)
	}
	if entity == nil || note.ID != id || entity.ID() != id {
		return nil, fmt.Errorf("journal identity does not match requested ID")
	}
	return entity, nil
}

// Update requires an existing journal and preserves its date and creation time.
// It re-encodes the document before refreshing the projection.
func (repository *JournalRepository) Update(ctx context.Context, entity *journals.Journal) error {
	if entity == nil || repository.index == nil {
		return fmt.Errorf("journal and repository dependencies are required")
	}
	existing, err := repository.ByID(ctx, entity.ID())
	if err != nil {
		return err
	}
	if !existing.Date().Equal(entity.Date()) || !existing.Metadata().CreatedAt().Equal(entity.Metadata().CreatedAt()) {
		return fmt.Errorf("journal date and creation time cannot change")
	}
	return repository.persist(ctx, entity)
}

func (repository *JournalRepository) persist(ctx context.Context, entity *journals.Journal) error {
	source, err := repository.codec.Encode(entity)
	if err != nil {
		return fmt.Errorf("encode journal: %w", err)
	}
	path, err := repository.notes.Put(ctx, ports.Note{ID: entity.ID(), Content: source})
	if err != nil {
		return fmt.Errorf("store journal: %w", err)
	}
	if err := repository.index.UpsertJournal(ctx, entity, path); err != nil {
		return fmt.Errorf("update journal index (document saved): %w", err)
	}
	return nil
}
