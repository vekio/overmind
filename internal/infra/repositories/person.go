package repositories

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/ports"
)

var _ ports.PersonRepository = (*PersonRepository)(nil)

type PersonRepository struct {
	notes ports.NoteStore
	index ports.Index
	codec ports.PersonCodec
}

func NewPersonRepository(notes ports.NoteStore, index ports.Index, codec ports.PersonCodec) *PersonRepository {
	return &PersonRepository{notes: notes, index: index, codec: codec}
}

func (repository *PersonRepository) Save(ctx context.Context, person *persons.Person) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if repository.notes == nil || repository.index == nil || repository.codec == nil {
		return fmt.Errorf("person repository dependencies are not configured")
	}
	source, err := repository.codec.Encode(person)
	if err != nil {
		return fmt.Errorf("encode person: %w", err)
	}
	path, err := repository.notes.Put(ctx, ports.Note{ID: person.ID(), Content: source})
	if err != nil {
		return fmt.Errorf("store person note: %w", err)
	}
	if err := repository.index.UpsertPerson(ctx, person, path); err != nil {
		return fmt.Errorf("update person index (document saved): %w", err)
	}
	return nil
}

func (repository *PersonRepository) ByID(ctx context.Context, id uuid.UUID) (*persons.Person, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id == uuid.Nil() {
		return nil, fmt.Errorf("person ID is required")
	}
	if repository.notes == nil || repository.index == nil || repository.codec == nil {
		return nil, fmt.Errorf("person repository dependencies are not configured")
	}
	note, err := repository.notes.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ports.ErrNoteNotFound) {
			return nil, fmt.Errorf("%w: %s", ports.ErrPersonNotFound, id)
		}
		return nil, fmt.Errorf("read person note: %w", err)
	}
	if note.ID != id {
		return nil, fmt.Errorf("stored note ID does not match requested ID")
	}
	person, err := repository.codec.Decode(note.Content)
	if err != nil {
		return nil, fmt.Errorf("map person note: %w", err)
	}
	if person.ID() != id {
		return nil, fmt.Errorf("person note ID does not match requested ID")
	}
	return person, nil
}
