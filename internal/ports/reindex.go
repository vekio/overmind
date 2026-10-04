package ports

import "context"

// NoteScanner visits documents independently of the current index.
type NoteScanner interface {
	Scan(context.Context, func(Note) error) error
}

// NoteProjector decodes a document and writes its validated projection.
type NoteProjector interface {
	Project(context.Context, Note, Index) error
}

// IndexRebuilder replaces all projections atomically. The supplied index is
// scoped to populate; it must not be retained after the callback returns.
type IndexRebuilder interface {
	Rebuild(context.Context, func(Index) error) error
}
