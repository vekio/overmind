// Package ports defines the infrastructure required by the application.
package ports

import (
	"context"

	"git.casta.me/alberto/overmind/internal/domain"
	"uuid"
)

// IDGenerator generates note identifiers.
type IDGenerator interface {
	Generate() uuid.UUID
}

// NoteRenderer renders a note as a document.
type NoteRenderer interface {
	Render(context.Context, domain.Note) ([]byte, error)
}

// NoteWriter persists a rendered note.
type NoteWriter interface {
	Write(context.Context, uuid.UUID, []byte) (string, error)
}

// NoteIndex indexes note metadata.
type NoteIndex interface {
	Upsert(context.Context, domain.Note) error
	JournalExists(context.Context, domain.Date) (bool, error)
}
