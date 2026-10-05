// Package persons models people and their group memberships.
package persons

import (
	"errors"
	"fmt"
	"github.com/vekio/overmind/internal/domain/shared"
	"time"
	"uuid"
)

// ErrInvalidPerson identifies an uninitialized entity or invalid required fields.
var ErrInvalidPerson = errors.New("invalid person")

// Person is an entity identified by UUID; edits preserve its identity and creation time.
type Person struct {
	id       uuid.UUID
	name     shared.Title
	groups   Groups
	tags     shared.Tags
	content  string
	metadata shared.EntityMetadata
}

// NewPerson validates identity and required domain values without assigning timestamps.
func NewPerson(id uuid.UUID, name shared.Title, groups Groups, tags shared.Tags, metadata shared.EntityMetadata) (*Person, error) {
	person := &Person{id: id, name: name, groups: groups, tags: tags, metadata: metadata}
	if err := person.validate(); err != nil {
		return nil, err
	}
	return person, nil
}
func (person *Person) validate() error {
	if person == nil || person.id == uuid.Nil() || person.name.IsZero() || person.metadata.IsZero() {
		return fmt.Errorf("%w: identity, name and metadata are required", ErrInvalidPerson)
	}
	return nil
}

// ID returns the stable document identity.
func (person *Person) ID() uuid.UUID { return person.id }

// Name returns the validated display name.
func (person *Person) Name() shared.Title { return person.name }

// Groups returns the immutable ordered membership collection.
func (person *Person) Groups() Groups { return person.groups }

// Tags returns the immutable ordered tag collection.
func (person *Person) Tags() shared.Tags { return person.tags }

// Metadata returns immutable creation and update timestamps.
func (person *Person) Metadata() shared.EntityMetadata { return person.metadata }

// Rename changes the display name after validating it and the update time.
// Failure leaves the entity unchanged.
func (person *Person) Rename(name shared.Title, at time.Time) error {
	if err := person.validate(); err != nil {
		return err
	}
	if name.IsZero() {
		return fmt.Errorf("%w: name is required", ErrInvalidPerson)
	}
	metadata, err := person.metadata.Updated(at)
	if err != nil {
		return err
	}
	person.name, person.metadata = name, metadata
	return nil
}

// Regroup replaces all memberships; an empty collection removes them.
func (person *Person) Regroup(groups Groups, at time.Time) error {
	if err := person.validate(); err != nil {
		return err
	}
	metadata, err := person.metadata.Updated(at)
	if err != nil {
		return err
	}
	person.groups, person.metadata = groups, metadata
	return nil
}

// ReplaceTags replaces all tags; an empty collection clears them.
// Invalid update times leave the entity unchanged.
func (person *Person) ReplaceTags(tags shared.Tags, at time.Time) error {
	if err := person.validate(); err != nil {
		return err
	}
	metadata, err := person.metadata.Updated(at)
	if err != nil {
		return err
	}
	person.tags, person.metadata = tags, metadata
	return nil
}

// Summary returns the display name.
func (person *Person) Summary() string { return person.name.String() }

// Content returns the original AsciiDoc body.
func (person *Person) Content() string { return person.content }

// Rewrite replaces the body verbatim, including empty content.
// Invalid update times leave content and metadata unchanged.
func (person *Person) Rewrite(content string, at time.Time) error {
	if err := person.validate(); err != nil {
		return err
	}
	metadata, err := person.metadata.Updated(at)
	if err != nil {
		return err
	}
	person.content, person.metadata = content, metadata
	return nil
}
