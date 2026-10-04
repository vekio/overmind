package bookmarks

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/shared"
)

var ErrInvalidBookmark = errors.New("invalid bookmark")

// Bookmark is a saved HTTP or HTTPS link, identified by a stable UUID.
// Mutations update metadata after validating the new values and time.
type Bookmark struct {
	id       uuid.UUID
	url      URL
	tags     shared.Tags
	metadata shared.EntityMetadata
}

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
func (entity *Bookmark) ID() uuid.UUID                   { return entity.id }
func (entity *Bookmark) Metadata() shared.EntityMetadata { return entity.metadata }
func (entity *Bookmark) Tags() shared.Tags               { return entity.tags }
func (entity *Bookmark) URL() URL                        { return entity.url }
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
func (entity *Bookmark) Summary() string { return entity.url.String() }
