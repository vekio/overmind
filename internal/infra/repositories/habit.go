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

type HabitRepository struct {
	notes ports.NoteStore
	index ports.Index
	codec ports.HabitCodec
}

func NewHabitRepository(notes ports.NoteStore, index ports.Index, codec ports.HabitCodec) *HabitRepository {
	return &HabitRepository{notes: notes, index: index, codec: codec}
}

func (repository *HabitRepository) Save(ctx context.Context, habit *habits.Habit) error {
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
	if habit.ID() != id {
		return nil, fmt.Errorf("habit note ID does not match requested ID")
	}
	return habit, nil
}
