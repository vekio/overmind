package domain

import (
	"errors"
	"fmt"
	"time"
	"uuid"
)

var ErrInvalidPage = errors.New("invalid page")

// Page is a titled note that may belong to an area.
type Page struct {
	metadata Metadata
	title    Title
	area     Area
}

// NewPage creates a page from validated values.
func NewPage(noteID uuid.UUID, title Title, area Area, tags Tags, createdAt time.Time) (Page, error) {
	if title.IsZero() {
		return Page{}, fmt.Errorf("%w: title must not be zero", ErrInvalidPage)
	}
	metadata, err := newMetadata(noteID, NoteKindPage, tags, createdAt, createdAt)
	if err != nil {
		return Page{}, fmt.Errorf("%w: %w", ErrInvalidPage, err)
	}

	return Page{
		metadata: metadata,
		title:    title,
		area:     area,
	}, nil
}

// Metadata returns the page metadata.
func (page Page) Metadata() Metadata {
	return page.metadata
}

// Title returns the page title.
func (page Page) Title() Title {
	return page.title
}

// Area returns the page area. It may be zero when the page has no area.
func (page Page) Area() Area {
	return page.area
}
