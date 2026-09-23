package domain

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidNoteKind indicates an unsupported note kind.
var ErrInvalidNoteKind = errors.New("invalid note kind")

// NoteKind identifies a supported note type.
type NoteKind string

const (
	NoteKindPage     NoteKind = "page"
	NoteKindJournal  NoteKind = "journal"
	NoteKindInbox    NoteKind = "inbox"
	NoteKindBookmark NoteKind = "bookmark"
)

// ParseNoteKind parses a note kind.
func ParseNoteKind(value string) (NoteKind, error) {
	kind := NoteKind(strings.TrimSpace(value))
	if !kind.IsValid() {
		return "", fmt.Errorf("%w: %q", ErrInvalidNoteKind, value)
	}
	return kind, nil
}

// IsValid reports whether kind is supported.
func (kind NoteKind) IsValid() bool {
	switch kind {
	case NoteKindPage, NoteKindJournal, NoteKindInbox, NoteKindBookmark:
		return true
	default:
		return false
	}
}

// String returns the serialized note kind.
func (kind NoteKind) String() string {
	return string(kind)
}
