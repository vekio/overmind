package pages

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/shared"
)

// ErrInvalidPage identifies an uninitialized entity or invalid required fields.
var ErrInvalidPage = errors.New("invalid page")

// Page is an entity with a stable UUID and explicit lifecycle metadata.
type Page struct {
	id       uuid.UUID
	title    shared.Title
	content  string
	area     Area
	tags     shared.Tags
	metadata shared.EntityMetadata
}

// NewPage validates identity and required domain values without assigning timestamps.
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

// ID returns the stable document identity.
func (entity *Page) ID() uuid.UUID { return entity.id }

// Title returns the validated page title.
func (entity *Page) Title() shared.Title { return entity.title }

// Tags returns the immutable ordered tag collection.
func (entity *Page) Tags() shared.Tags { return entity.tags }

// Metadata returns immutable creation and update timestamps.
func (entity *Page) Metadata() shared.EntityMetadata { return entity.metadata }

// ReplaceTags replaces all tags; an empty collection clears them.
// Invalid update times leave the entity unchanged.
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

// Area returns the hierarchical location; its zero value means no area.
func (entity *Page) Area() Area { return entity.area }

// Rename changes the title after validating it and the update time.
// Failure leaves the entity unchanged.
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

// Summary returns the page title for display.
func (entity *Page) Summary() string { return entity.title.String() }

// Content returns the original AsciiDoc body.
func (entity *Page) Content() string { return entity.content }

// Rewrite replaces the body verbatim, including empty content.
// Invalid update times leave content and metadata unchanged.
func (entity *Page) Rewrite(content string, at time.Time) error {
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
