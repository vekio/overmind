package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"uuid"
)

// ErrInvalidInbox indicates an invalid inbox note.
var ErrInvalidInbox = errors.New("invalid inbox")

// Inbox is a quick note awaiting review.
type Inbox struct {
	metadata Metadata
	content  string
}

func (Inbox) isNote() {}

// NewInbox creates an inbox note from validated content.
func NewInbox(noteID uuid.UUID, content string, createdAt time.Time) (Inbox, error) {
	if strings.TrimSpace(content) == "" {
		return Inbox{}, fmt.Errorf("%w: content must not be empty", ErrInvalidInbox)
	}
	metadata, err := newMetadata(noteID, NoteKindInbox, Tags{}, createdAt, createdAt)
	if err != nil {
		return Inbox{}, fmt.Errorf("%w: %w", ErrInvalidInbox, err)
	}

	return Inbox{
		metadata: metadata,
		content:  content,
	}, nil
}

// Metadata returns the inbox metadata.
func (inbox Inbox) Metadata() Metadata {
	return inbox.metadata
}

// Content returns the captured content.
func (inbox Inbox) Content() string {
	return inbox.content
}
