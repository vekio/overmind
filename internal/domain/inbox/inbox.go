package inbox

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/shared"
)

// ErrInvalidInbox identifies an uninitialized entity or invalid required fields.
var ErrInvalidInbox = errors.New("invalid inbox")

// Inbox is an unstructured note; its content may be empty, identified by a stable UUID.
// Mutations update metadata after validating the new values and time.
type Inbox struct {
	id       uuid.UUID
	content  string
	tags     shared.Tags
	metadata shared.EntityMetadata
}

// NewInbox validates identity and required domain values without assigning timestamps.
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

// ID returns the stable document identity.
func (entity *Inbox) ID() uuid.UUID { return entity.id }

// Metadata returns immutable creation and update timestamps.
func (entity *Inbox) Metadata() shared.EntityMetadata { return entity.metadata }

// Tags returns the immutable ordered tag collection.
func (entity *Inbox) Tags() shared.Tags { return entity.tags }

// Content returns the document body without parsing or normalizing it.
func (entity *Inbox) Content() string { return entity.content }

// ReplaceTags replaces all tags; an empty collection clears them.
// Invalid update times leave the entity unchanged.
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

// Rewrite replaces the body verbatim, including empty content.
// Invalid update times leave content and metadata unchanged.
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
