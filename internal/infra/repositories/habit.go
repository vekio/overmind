package repositories

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/ports"
)

var _ ports.HabitRepository = (*HabitRepository)(nil)

// HabitRepository stores managed habit documents and their derived index entries.
type HabitRepository struct {
	notes ports.NoteStore
	index ports.Index
	codec ports.HabitCodec
}

// NewHabitRepository binds source storage, index and typed document codec.
func NewHabitRepository(notes ports.NoteStore, index ports.Index, codec ports.HabitCodec) *HabitRepository {
	return &HabitRepository{notes: notes, index: index, codec: codec}
}

// Save writes the habit document before updating its index projection.
// An indexing error can occur after the document has been saved.
func (repository *HabitRepository) Save(ctx context.Context, habit *habits.Habit) error {
	return repository.persist(ctx, habit)
}

func (repository *HabitRepository) persist(ctx context.Context, habit *habits.Habit) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if repository.notes == nil || repository.index == nil || repository.codec == nil {
		return fmt.Errorf("habit repository dependencies are not configured")
	}
	source, err := repository.codec.Encode(habit)
	if err != nil {
		return fmt.Errorf("encode habit: %w", err)
	}
	path, err := repository.notes.Put(ctx, ports.Note{ID: habit.ID(), Content: source})
	if err != nil {
		return fmt.Errorf("store habit note: %w", err)
	}
	if err := repository.index.UpsertHabit(ctx, habit, path); err != nil {
		return fmt.Errorf("update habit index (document saved): %w", err)
	}
	return nil
}

// ByID decodes the authoritative source and verifies its habit kind and UUID.
func (repository *HabitRepository) ByID(ctx context.Context, id uuid.UUID) (*habits.Habit, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id == uuid.Nil() {
		return nil, fmt.Errorf("habit ID is required")
	}
	if repository.notes == nil || repository.index == nil || repository.codec == nil {
		return nil, fmt.Errorf("habit repository dependencies are not configured")
	}
	note, err := repository.notes.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ports.ErrNoteNotFound) {
			return nil, fmt.Errorf("%w: %s", ports.ErrHabitNotFound, id)
		}
		return nil, fmt.Errorf("read habit note: %w", err)
	}
	if note.ID != id {
		return nil, fmt.Errorf("stored note ID does not match requested ID")
	}
	habit, err := repository.codec.Decode(note.Content)
	if err != nil {
		return nil, fmt.Errorf("map habit note: %w", err)
	}
	if habit == nil || habit.ID() != id {
		return nil, fmt.Errorf("habit note ID does not match requested ID")
	}
	return habit, nil
}

// Update requires an existing habit and preserves its creation time.
// It re-encodes the document before refreshing the projection.
func (repository *HabitRepository) Update(ctx context.Context, entity *habits.Habit) error {
	if entity == nil || repository.index == nil {
		return fmt.Errorf("habit and repository dependencies are required")
	}
	existing, err := repository.ByID(ctx, entity.ID())
	if err != nil {
		return err
	}
	if !existing.Metadata().CreatedAt().Equal(entity.Metadata().CreatedAt()) {
		return fmt.Errorf("habit creation time cannot change")
	}
	return repository.persist(ctx, entity)
}
