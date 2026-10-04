// Package persons models people and their group memberships.
package persons

import (
	"errors"
	"fmt"
	"github.com/vekio/overmind/internal/domain/shared"
	"time"
	"uuid"
)

var ErrInvalidPerson = errors.New("invalid person")

// Person is an entity identified by UUID; edits preserve its identity and creation time.
type Person struct {
	id       uuid.UUID
	name     shared.Title
	groups   Groups
	tags     shared.Tags
	metadata shared.EntityMetadata
}

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
func (person *Person) ID() uuid.UUID                   { return person.id }
func (person *Person) Name() shared.Title              { return person.name }
func (person *Person) Groups() Groups                  { return person.groups }
func (person *Person) Tags() shared.Tags               { return person.tags }
func (person *Person) Metadata() shared.EntityMetadata { return person.metadata }

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
func (person *Person) Summary() string { return person.name.String() }
