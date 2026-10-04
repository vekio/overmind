package inbox

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/shared"
)

var ErrInvalidInbox = errors.New("invalid inbox")

// Inbox is an unstructured note; its content may be empty, identified by a stable UUID.
// Mutations update metadata after validating the new values and time.
type Inbox struct {
	id       uuid.UUID
	content  string
	tags     shared.Tags
	metadata shared.EntityMetadata
}

func NewInbox(id uuid.UUID, content string, tags shared.Tags, metadata shared.EntityMetadata) (*Inbox, error) {
	entity := &Inbox{id: id, content: content, tags: tags, metadata: metadata}
	if err := entity.validate(); err != nil {
		return nil, err
	}
	return entity, nil
}
func (entity *Inbox) validate() error {
	if entity == nil || entity.id == uuid.Nil() || entity.metadata.IsZero() {
		return fmt.Errorf("%w: identity and metadata are required", ErrInvalidInbox)
	}
	return nil
}
func (entity *Inbox) ID() uuid.UUID                   { return entity.id }
func (entity *Inbox) Metadata() shared.EntityMetadata { return entity.metadata }
func (entity *Inbox) Tags() shared.Tags               { return entity.tags }
func (entity *Inbox) Content() string                 { return entity.content }
func (entity *Inbox) ReplaceTags(tags shared.Tags, at time.Time) error {
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
func (entity *Inbox) Rewrite(content string, at time.Time) error {
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
