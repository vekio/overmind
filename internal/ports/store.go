// Package ports defines the document, indexing and codec contracts used by the migrated notes.
package ports

import (
	"context"
	"errors"
	"uuid"
)

var ErrNoteNotFound = errors.New("note not found")

// Note is a stored document; its content format is determined by its codec.
type Note struct {
	ID      uuid.UUID
	Content []byte
	Path    string
}

// NoteStore persists documents independently of their domain model or format.
// Put returns the actual stored path; Get includes that path in the note.
type NoteStore interface {
	Put(context.Context, Note) (string, error)
	Get(context.Context, uuid.UUID) (Note, error)
	Delete(context.Context, uuid.UUID) error
}

// NoteKind identifies the persisted document format, not an entity property.
type NoteKind string

const (
	NoteKindPage     NoteKind = "page"
	NoteKindJournal  NoteKind = "journal"
	NoteKindHabit    NoteKind = "habit"
	NoteKindPerson   NoteKind = "person"
	NoteKindBookmark NoteKind = "bookmark"
	NoteKindInbox    NoteKind = "inbox"
)

func (kind NoteKind) String() string { return string(kind) }
