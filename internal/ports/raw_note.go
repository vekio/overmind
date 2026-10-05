package ports

import (
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/shared"
)

// RawNoteHeader identifies a note and its protected metadata.
type RawNoteHeader struct {
	ID       uuid.UUID
	Kind     NoteKind
	Metadata shared.EntityMetadata
	// Date is the fixed calendar date for journals and empty for other kinds.
	Date string
}

// RawNoteCodec validates complete AsciiDoc source without regenerating it.
type RawNoteCodec interface {
	// Inspect validates identity and metadata using only the header, allowing
	// documents with invalid bodies to be opened for repair.
	Inspect([]byte) (RawNoteHeader, error)

	// Stamp replaces only the effective header update-time value, preserving all
	// other bytes. Callers validate the resulting document before persistence.
	Stamp([]byte, time.Time) ([]byte, error)

	// Decode validates supported AsciiDoc syntax and the typed domain fields. A nonempty
	// expected kind must match the header; an empty kind accepts any supported kind.
	Decode(NoteKind, []byte) (RawNoteHeader, error)
}
