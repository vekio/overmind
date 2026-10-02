// Package ports defines the infrastructure required by the application.
package ports

import (
	"context"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain"
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

// NoteReader reads a note document by its identifier.
type NoteReader interface {
	Read(context.Context, uuid.UUID) ([]byte, error)
}

// NoteDeleter removes a note document by its identifier.
type NoteDeleter interface {
	Delete(context.Context, uuid.UUID) error
}

// NoteIndex indexes note metadata.
type NoteIndex interface {
	Upsert(context.Context, domain.Note) error
	UpsertRecord(context.Context, IndexRecord) error
	Delete(context.Context, uuid.UUID) error
	JournalExists(context.Context, domain.Date) (bool, error)
	JournalID(context.Context, domain.Date) (uuid.UUID, bool, error)
	ReplaceAll(context.Context, []IndexRecord) error
}

// NoteLister reads indexed note metadata without depending on a presentation.
type NoteLister interface {
	List(context.Context) ([]IndexRecord, error)
}

// NoteWalker visits AsciiDoc note files and supplies their content.
type NoteWalker interface {
	Walk(context.Context, func(path string, content []byte) error) error
}

// NoteParser recovers indexable metadata from one document.
type NoteParser interface {
	Parse([]byte) (IndexRecord, error)
}

// IndexRecord contains metadata recovered from one note document.
type IndexRecord struct {
	ID         uuid.UUID
	Kind       domain.NoteKind
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Attributes []IndexAttribute
	Tags       []string
}

// IndexAttribute is a searchable name-value pair associated with a note.
type IndexAttribute struct {
	Name  string
	Value string
}
