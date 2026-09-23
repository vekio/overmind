package domain

import (
	"errors"
	"fmt"
	"time"
	"uuid"
)

// ErrInvalidBookmark indicates an invalid bookmark note.
var ErrInvalidBookmark = errors.New("invalid bookmark")

// Bookmark is a saved HTTP or HTTPS link.
type Bookmark struct {
	metadata Metadata
	url      URL
}

func (Bookmark) isNote() {}

// NewBookmark creates a bookmark from validated values.
func NewBookmark(noteID uuid.UUID, url URL, tags Tags, createdAt time.Time) (Bookmark, error) {
	if url.IsZero() {
		return Bookmark{}, fmt.Errorf("%w: URL must not be zero", ErrInvalidBookmark)
	}
	metadata, err := newMetadata(noteID, NoteKindBookmark, tags, createdAt, createdAt)
	if err != nil {
		return Bookmark{}, fmt.Errorf("%w: %w", ErrInvalidBookmark, err)
	}

	return Bookmark{
		metadata: metadata,
		url:      url,
	}, nil
}

// Metadata returns the bookmark metadata.
func (bookmark Bookmark) Metadata() Metadata {
	return bookmark.metadata
}

// URL returns the bookmark URL.
func (bookmark Bookmark) URL() URL {
	return bookmark.url
}
