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

// PersonRepository stores managed person documents and their derived index entries.
type PersonRepository struct {
	notes ports.NoteStore
	index ports.Index
	codec ports.PersonCodec
}

// NewPersonRepository binds source storage, index and typed document codec.
func NewPersonRepository(notes ports.NoteStore, index ports.Index, codec ports.PersonCodec) *PersonRepository {
	return &PersonRepository{notes: notes, index: index, codec: codec}
}

// Save writes the person document before updating its index projection.
// An indexing error can occur after the document has been saved.
func (repository *PersonRepository) Save(ctx context.Context, person *persons.Person) error {
	return repository.persist(ctx, person)
}

func (repository *PersonRepository) persist(ctx context.Context, person *persons.Person) error {
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

// ByID decodes the authoritative source and verifies its person kind and UUID.
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
	if person == nil || person.ID() != id {
		return nil, fmt.Errorf("person note ID does not match requested ID")
	}
	return person, nil
}

// Update requires an existing person and preserves its creation time.
// It re-encodes the document before refreshing the projection.
func (repository *PersonRepository) Update(ctx context.Context, entity *persons.Person) error {
	if entity == nil || repository.index == nil {
		return fmt.Errorf("person and repository dependencies are required")
	}
	existing, err := repository.ByID(ctx, entity.ID())
	if err != nil {
		return err
	}
	if !existing.Metadata().CreatedAt().Equal(entity.Metadata().CreatedAt()) {
		return fmt.Errorf("person creation time cannot change")
	}
	return repository.persist(ctx, entity)
}
