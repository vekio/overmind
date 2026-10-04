package persons_test

import (
	"errors"
	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/domain/shared"
	"reflect"
	"testing"
	"time"
	"uuid"
)

func TestPersonIdentityEditsAndMetadata(t *testing.T) {
	at := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	name, _ := shared.NewTitle("Ana García")
	metadata, _ := shared.NewEntityMetadata(at, at)
	person, err := persons.NewPerson(uuid.New(), name, persons.Groups{}, shared.Tags{}, metadata)
	if err != nil {
		t.Fatal(err)
	}
	alias, id := person, person.ID()
	renamed, _ := shared.NewTitle("Ana García López")
	work, _ := persons.NewGroup("Trabajo")
	groups, _ := persons.NewGroups(work)
	tag, _ := shared.NewTag("Amiga")
	tags, _ := shared.NewTags(tag)
	for step, edit := range []func(time.Time) error{
		func(at time.Time) error { return person.Rename(renamed, at) },
		func(at time.Time) error { return person.Regroup(groups, at) },
		func(at time.Time) error { return person.ReplaceTags(tags, at) },
	} {
		changedAt := at.Add(time.Duration(step+1) * time.Hour)
		if err := edit(changedAt); err != nil {
			t.Fatal(err)
		}
		if alias.ID() != id || !alias.Metadata().CreatedAt().Equal(at) || !alias.Metadata().UpdatedAt().Equal(changedAt) {
			t.Fatal("edit changed identity or creation time")
		}
	}
	if person.Name() != renamed || !person.Groups().Contains(work) || !person.Tags().Contains(tag) {
		t.Fatal("edit did not mutate the entity")
	}
	before := *person
	for _, edit := range []func() error{
		func() error { return person.Rename(shared.Title{}, at.Add(4*time.Hour)) },
		func() error { return person.Regroup(persons.Groups{}, at) },
		func() error { return person.ReplaceTags(shared.Tags{}, time.Time{}) },
	} {
		if err := edit(); err == nil || !reflect.DeepEqual(*person, before) {
			t.Fatal("invalid edit partially modified person")
		}
	}
	if err := person.Regroup(persons.Groups{}, at.Add(4*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if !person.Groups().IsEmpty() {
		t.Fatal("empty groups must remove memberships")
	}
}

func TestPersonRequiresIdentityNameAndMetadata(t *testing.T) {
	name, _ := shared.NewTitle("Ana")
	at := time.Now()
	metadata, _ := shared.NewEntityMetadata(at, at)
	for _, input := range []struct {
		id       uuid.UUID
		name     shared.Title
		metadata shared.EntityMetadata
	}{
		{uuid.Nil(), name, metadata}, {uuid.New(), shared.Title{}, metadata}, {uuid.New(), name, shared.EntityMetadata{}},
	} {
		person, err := persons.NewPerson(input.id, input.name, persons.Groups{}, shared.Tags{}, input.metadata)
		if !errors.Is(err, persons.ErrInvalidPerson) || person != nil {
			t.Fatalf("invalid person = %v", err)
		}
	}
	for _, person := range []*persons.Person{nil, {}} {
		if err := person.Rename(name, at); err == nil {
			t.Fatal("uninitialized person renamed")
		}
		if err := person.Regroup(persons.Groups{}, at); err == nil {
			t.Fatal("uninitialized person regrouped")
		}
		if err := person.ReplaceTags(shared.Tags{}, at); err == nil {
			t.Fatal("uninitialized person tagged")
		}
	}
}
