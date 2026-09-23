package domain

import (
	"errors"
	"fmt"
	"time"
	"uuid"
)

// ErrInvalidMetadata indicates invalid common note metadata.
var ErrInvalidMetadata = errors.New("invalid note metadata")

// Metadata contains the properties shared by every note.
type Metadata struct {
	id        uuid.UUID
	kind      NoteKind
	tags      Tags
	createdAt time.Time
	updatedAt time.Time
}

func newMetadata(
	id uuid.UUID,
	kind NoteKind,
	tags Tags,
	createdAt time.Time,
	updatedAt time.Time,
) (Metadata, error) {
	if id == uuid.Nil() {
		return Metadata{}, fmt.Errorf("%w: UUID must not be nil", ErrInvalidMetadata)
	}
	if !kind.IsValid() {
		return Metadata{}, fmt.Errorf("%w: %w", ErrInvalidMetadata, ErrInvalidNoteKind)
	}
	if createdAt.IsZero() {
		return Metadata{}, fmt.Errorf("%w: creation time must not be zero", ErrInvalidMetadata)
	}
	if updatedAt.IsZero() {
		return Metadata{}, fmt.Errorf("%w: update time must not be zero", ErrInvalidMetadata)
	}
	if updatedAt.Before(createdAt) {
		return Metadata{}, fmt.Errorf("%w: update time precedes creation time", ErrInvalidMetadata)
	}

	return Metadata{
		id:        id,
		kind:      kind,
		tags:      tags,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}, nil
}

// ID returns the note identifier.
func (metadata Metadata) ID() uuid.UUID {
	return metadata.id
}

// Kind returns the note kind.
func (metadata Metadata) Kind() NoteKind {
	return metadata.kind
}

// Tags returns the note tags.
func (metadata Metadata) Tags() Tags {
	return metadata.tags
}

// CreatedAt returns the note creation time.
func (metadata Metadata) CreatedAt() time.Time {
	return metadata.createdAt
}

// UpdatedAt returns the last note update time.
func (metadata Metadata) UpdatedAt() time.Time {
	return metadata.updatedAt
}
