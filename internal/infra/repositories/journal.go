package repositories

import (
	"context"
	"fmt"

	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/ports"
)

var _ ports.JournalRepository = (*JournalRepository)(nil)

type JournalRepository struct {
	notes ports.NoteStore
	index ports.Index
	codec ports.JournalEncoder
}

func NewJournalRepository(notes ports.NoteStore, index ports.Index, codec ports.JournalEncoder) *JournalRepository {
	return &JournalRepository{notes: notes, index: index, codec: codec}
}
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
