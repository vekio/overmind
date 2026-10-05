package bookmarks

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/shared"
)

// ErrInvalidBookmark identifies an uninitialized entity or invalid required fields.
var ErrInvalidBookmark = errors.New("invalid bookmark")

// Bookmark is a saved HTTP or HTTPS link, identified by a stable UUID.
// Mutations update metadata after validating the new values and time.
type Bookmark struct {
	id       uuid.UUID
	url      URL
	tags     shared.Tags
	content  string
	metadata shared.EntityMetadata
}

// NewBookmark validates identity and required domain values without assigning timestamps.
func NewBookmark(id uuid.UUID, url URL, tags shared.Tags, metadata shared.EntityMetadata) (*Bookmark, error) {
	entity := &Bookmark{id: id, url: url, tags: tags, metadata: metadata}
	if err := entity.validate(); err != nil {
		return nil, err
	}
	return entity, nil
}
func (entity *Bookmark) validate() error {
	if entity == nil || entity.id == uuid.Nil() || entity.metadata.IsZero() {
		return fmt.Errorf("%w: identity and metadata are required", ErrInvalidBookmark)
	}
	if entity.url.IsZero() {
		return fmt.Errorf("%w: URL is required", ErrInvalidBookmark)
	}
	return nil
}

// ID returns the stable document identity.
func (entity *Bookmark) ID() uuid.UUID { return entity.id }

// Metadata returns immutable creation and update timestamps.
func (entity *Bookmark) Metadata() shared.EntityMetadata { return entity.metadata }

// Tags returns the immutable ordered tag collection.
func (entity *Bookmark) Tags() shared.Tags { return entity.tags }

// URL returns the validated HTTP or HTTPS link.
func (entity *Bookmark) URL() URL { return entity.url }

// ReplaceTags replaces all tags; an empty collection clears them.
// Invalid update times leave the entity unchanged.
func (entity *Bookmark) ReplaceTags(tags shared.Tags, at time.Time) error {
	if err := entity.validate(); err != nil {
		return err
	}
	metadata, err := entity.metadata.Updated(at)
	if err != nil {
		return err
	}
	entity.tags, entity.metadata = tags, metadata
	return nil
}

// ChangeURL replaces the link after validating it and the update time.
// Failure leaves both the link and metadata unchanged.
func (entity *Bookmark) ChangeURL(url URL, at time.Time) error {
	if err := entity.validate(); err != nil {
		return err
	}
	if url.IsZero() {
		return fmt.Errorf("%w: URL is required", ErrInvalidBookmark)
	}
	metadata, err := entity.metadata.Updated(at)
	if err != nil {
		return err
	}
	entity.url, entity.metadata = url, metadata
	return nil
}

// Summary returns the link text for display.
func (entity *Bookmark) Summary() string { return entity.url.String() }

// Content returns the original AsciiDoc body.
func (entity *Bookmark) Content() string { return entity.content }

// Rewrite replaces the body verbatim, including empty content.
// Invalid update times leave content and metadata unchanged.
func (entity *Bookmark) Rewrite(content string, at time.Time) error {
	if err := entity.validate(); err != nil {
		return err
	}
	metadata, err := entity.metadata.Updated(at)
	if err != nil {
		return err
	}
	entity.content, entity.metadata = content, metadata
	return nil
}
