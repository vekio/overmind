package journals

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/shared"
)

// ErrInvalidJournal identifies an uninitialized entity or invalid required fields.
var ErrInvalidJournal = errors.New("invalid journal")

// Journal associates content with a fixed date and stable UUID.
// Moving an entry to another day requires creating a new journal.
type Journal struct {
	id       uuid.UUID
	date     calendar.Date
	content  string
	tags     shared.Tags
	metadata shared.EntityMetadata
}

// NewJournal validates identity and required domain values without assigning timestamps.
func NewJournal(id uuid.UUID, date calendar.Date, content string, tags shared.Tags, metadata shared.EntityMetadata) (*Journal, error) {
	entity := &Journal{id: id, date: date, content: content, tags: tags, metadata: metadata}
	if err := entity.validate(); err != nil {
		return nil, err
	}
	return entity, nil
}
func (entity *Journal) validate() error {
	if entity == nil || entity.id == uuid.Nil() || entity.date.IsZero() || entity.metadata.IsZero() {
		return fmt.Errorf("%w: identity, date and metadata are required", ErrInvalidJournal)
	}
	return nil
}

// ID returns the stable document identity.
func (entity *Journal) ID() uuid.UUID { return entity.id }

// Date returns the fixed calendar date associated with this entry.
func (entity *Journal) Date() calendar.Date { return entity.date }

// Content returns the document body without parsing or normalizing it.
func (entity *Journal) Content() string { return entity.content }

// Tags returns the immutable ordered tag collection.
func (entity *Journal) Tags() shared.Tags { return entity.tags }

// Metadata returns immutable creation and update timestamps.
func (entity *Journal) Metadata() shared.EntityMetadata { return entity.metadata }

// ReplaceTags replaces all tags; an empty collection clears them.
// Invalid update times leave the entity unchanged.
func (entity *Journal) ReplaceTags(tags shared.Tags, at time.Time) error {
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

// Rewrite replaces the body verbatim, including empty content.
// Invalid update times leave content and metadata unchanged.
func (entity *Journal) Rewrite(content string, at time.Time) error {
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

// Summary returns a display label derived from the fixed journal date.
func (entity *Journal) Summary() string { return "Journal " + entity.date.String() }
