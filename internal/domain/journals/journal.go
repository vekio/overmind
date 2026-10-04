package journals

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/shared"
)

var ErrInvalidJournal = errors.New("invalid journal")

// Journal is an entity with a stable UUID and explicit lifecycle metadata.
type Journal struct {
	id       uuid.UUID
	date     calendar.Date
	tags     shared.Tags
	metadata shared.EntityMetadata
}

func NewJournal(id uuid.UUID, date calendar.Date, tags shared.Tags, metadata shared.EntityMetadata) (*Journal, error) {
	entity := &Journal{id: id, date: date, tags: tags, metadata: metadata}
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
func (entity *Journal) ID() uuid.UUID                   { return entity.id }
func (entity *Journal) Date() calendar.Date             { return entity.date }
func (entity *Journal) Tags() shared.Tags               { return entity.tags }
func (entity *Journal) Metadata() shared.EntityMetadata { return entity.metadata }
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

// Date stays fixed: moving a journal to another day requires creating a new entry.
func (entity *Journal) Summary() string { return "Journal " + entity.date.String() }
