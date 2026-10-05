// Package ports defines the storage, indexing and codec contracts used by application handlers.
package ports

import (
	"context"
	"errors"
	"uuid"
)

// ErrNoteNotFound identifies a missing source document.
var ErrNoteNotFound = errors.New("note not found")

// Note carries the complete source bytes and storage location of a managed document.
type Note struct {
	ID      uuid.UUID
	Content []byte
	Path    string
}

// NoteStore owns source documents independently of their domain model and index.
// Implementations preserve supplied bytes and do not perform domain validation.
type NoteStore interface {
	// Scan visits managed documents in stable path order without consulting the
	// index. Cancellation or a visitor error stops the scan.
	Scan(context.Context, func(Note) error) error

	// Put atomically creates or replaces the document identified by Note.ID and
	// returns its storage path. Updating its index is a separate operation.
	Put(context.Context, Note) (string, error)

	// Get reads the complete document and its path, or returns ErrNoteNotFound.
	Get(context.Context, uuid.UUID) (Note, error)

	// Delete removes the source document. Missing documents are treated as already
	// deleted; ambiguous UUIDs must be rejected before removing any file.
	Delete(context.Context, uuid.UUID) error
}

// NoteKind identifies the persisted document format, not an entity property.
type NoteKind string

const (
	// NoteKindPage identifies managed page documents.
	NoteKindPage NoteKind = "page"
	// NoteKindJournal identifies managed journal documents.
	NoteKindJournal NoteKind = "journal"
	// NoteKindHabit identifies managed habit documents.
	NoteKindHabit NoteKind = "habit"
	// NoteKindPerson identifies managed person documents.
	NoteKindPerson NoteKind = "person"
	// NoteKindBookmark identifies managed bookmark documents.
	NoteKindBookmark NoteKind = "bookmark"
	// NoteKindInbox identifies managed inbox documents.
	NoteKindInbox NoteKind = "inbox"
)

// String returns the document kind stored in managed headers and index projections.
func (kind NoteKind) String() string { return string(kind) }
