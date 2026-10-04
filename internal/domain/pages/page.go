package pages

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/shared"
)

var ErrInvalidPage = errors.New("invalid page")

// Page is an entity with a stable UUID and explicit lifecycle metadata.
type Page struct {
	id       uuid.UUID
	title    shared.Title
	area     Area
	tags     shared.Tags
	metadata shared.EntityMetadata
}

func NewPage(id uuid.UUID, title shared.Title, area Area, tags shared.Tags, metadata shared.EntityMetadata) (*Page, error) {
	entity := &Page{id: id, title: title, area: area, tags: tags, metadata: metadata}
	if err := entity.validate(); err != nil {
		return nil, err
	}
	return entity, nil
}
func (entity *Page) validate() error {
	if entity == nil || entity.id == uuid.Nil() || entity.title.IsZero() || entity.metadata.IsZero() {
		return fmt.Errorf("%w: identity, title and metadata are required", ErrInvalidPage)
	}
	return nil
}
func (entity *Page) ID() uuid.UUID                   { return entity.id }
func (entity *Page) Title() shared.Title             { return entity.title }
func (entity *Page) Tags() shared.Tags               { return entity.tags }
func (entity *Page) Metadata() shared.EntityMetadata { return entity.metadata }
func (entity *Page) ReplaceTags(tags shared.Tags, at time.Time) error {
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
func (entity *Page) Area() Area { return entity.area }
func (entity *Page) Rename(title shared.Title, at time.Time) error {
	if err := entity.validate(); err != nil {
		return err
	}
	if title.IsZero() {
		return fmt.Errorf("%w: title is required", ErrInvalidPage)
	}
	metadata, err := entity.metadata.Updated(at)
	if err != nil {
		return err
	}
	entity.title, entity.metadata = title, metadata
	return nil
}

// MoveTo changes the page's area; a zero area removes its location.
func (entity *Page) MoveTo(area Area, at time.Time) error {
	if err := entity.validate(); err != nil {
		return err
	}
	metadata, err := entity.metadata.Updated(at)
	if err != nil {
		return err
	}
	entity.area, entity.metadata = area, metadata
	return nil
}
func (entity *Page) Summary() string { return entity.title.String() }
