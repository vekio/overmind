package domain

import (
	"errors"
	"fmt"
	"time"
	"uuid"
)

var ErrInvalidPerson = errors.New("invalid person")

// Person is a note about someone the user knows.
type Person struct {
	metadata Metadata
	name     Title
	groups   Groups
}

func (Person) isNote() {}

// NewPerson creates a person from validated values.
func NewPerson(noteID uuid.UUID, name Title, groups Groups, tags Tags, createdAt time.Time) (Person, error) {
	if name.IsZero() {
		return Person{}, fmt.Errorf("%w: name must not be zero", ErrInvalidPerson)
	}
	metadata, err := newMetadata(noteID, NoteKindPerson, tags, createdAt, createdAt)
	if err != nil {
		return Person{}, fmt.Errorf("%w: %w", ErrInvalidPerson, err)
	}

	return Person{
		metadata: metadata,
		name:     name,
		groups:   groups,
	}, nil
}

// Metadata returns the person metadata.
func (person Person) Metadata() Metadata {
	return person.metadata
}

// Name returns the person's full name.
func (person Person) Name() Title {
	return person.name
}

// Groups returns the groups the person belongs to.
func (person Person) Groups() Groups {
	return person.groups
}
