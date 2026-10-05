package index_test

import (
	"context"
	"reflect"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/domain/shared"
	noteindex "github.com/vekio/overmind/internal/infra/index"
)

func TestFindGroupsOnlyListsUsedGroups(t *testing.T) {
	ctx := context.Background()
	index, err := noteindex.New(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	newPerson := func(names ...string) *persons.Person {
		var values []persons.Group
		for _, name := range names {
			group, err := persons.NewGroup(name)
			if err != nil {
				t.Fatal(err)
			}
			values = append(values, group)
		}
		groups, err := persons.NewGroups(values...)
		if err != nil {
			t.Fatal(err)
		}
		name, _ := shared.NewTitle("Person")
		entity, err := persons.NewPerson(uuid.New(), name, groups, shared.Tags{}, fixtureMetadata())
		if err != nil {
			t.Fatal(err)
		}
		return entity
	}
	first, second := newPerson("Work", "Friends"), newPerson("Work")
	for _, entity := range []*persons.Person{first, second} {
		if err := index.UpsertPerson(ctx, entity, "/vault/person.adoc"); err != nil {
			t.Fatal(err)
		}
	}
	assertGroups := func(want []string) {
		t.Helper()
		got, err := index.FindGroups(ctx)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("groups=%v, error=%v, want=%v", got, err, want)
		}
	}
	assertGroups([]string{"friends", "work"})
	if err := first.Regroup(persons.Groups{}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := index.UpsertPerson(ctx, first, "/vault/person.adoc"); err != nil {
		t.Fatal(err)
	}
	assertGroups([]string{"work"})
	if err := index.Delete(ctx, second.ID()); err != nil {
		t.Fatal(err)
	}
	got, err := index.FindGroups(ctx)
	if err != nil || len(got) != 0 {
		t.Fatalf("unused groups remain visible: %v, %v", got, err)
	}
}
