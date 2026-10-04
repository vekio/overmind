package ports

import (
	"context"
	"time"
	"uuid"
)

// NoteFilter selects indexed notes. Empty type and tag match all notes.
type NoteFilter struct {
	Type   string
	Tag    string
	Limit  int
	Offset int
}

// NoteSummary is a read projection, independent of the note's domain entity.
type NoteSummary struct {
	ID        uuid.UUID
	Type      string
	Label     string
	Tags      []string
	UpdatedAt time.Time
}

// NoteFinder reads summaries without loading or decoding note documents.
type NoteFinder interface {
	FindNotes(context.Context, NoteFilter) ([]NoteSummary, error)
}
