package domain

import (
	"errors"
	"fmt"
	"time"
	"uuid"
)

// ErrInvalidJournal indicates an invalid journal note.
var ErrInvalidJournal = errors.New("invalid journal")

// Journal is a note associated with a calendar date.
type Journal struct {
	metadata Metadata
	date     Date
}

func (Journal) isNote() {}

// NewJournal creates a journal from validated values.
func NewJournal(noteID uuid.UUID, date Date, tags Tags, createdAt time.Time) (Journal, error) {
	if date.IsZero() {
		return Journal{}, fmt.Errorf("%w: date must not be zero", ErrInvalidJournal)
	}
	metadata, err := newMetadata(noteID, NoteKindJournal, tags, createdAt, createdAt)
	if err != nil {
		return Journal{}, fmt.Errorf("%w: %w", ErrInvalidJournal, err)
	}

	return Journal{
		metadata: metadata,
		date:     date,
	}, nil
}

// Metadata returns the journal metadata.
func (journal Journal) Metadata() Metadata {
	return journal.metadata
}

// Date returns the journal date.
func (journal Journal) Date() Date {
	return journal.date
}
