package persons_test

import (
	"errors"
	"github.com/vekio/overmind/internal/domain/persons"
	"testing"
)

func TestGroupsAreNormalizedUniqueAndImmutable(t *testing.T) {
	work, err := persons.NewGroup(" Work colleagues ")
	if err != nil || work.String() != "work-colleagues" {
		t.Fatalf("normalized group = %q, %v", work, err)
	}
	if _, err := persons.NewGroup(" !!! "); !errors.Is(err, persons.ErrInvalidGroup) {
		t.Fatalf("invalid group = %v", err)
	}
	if _, err := persons.NewGroups(persons.Group{}); !errors.Is(err, persons.ErrInvalidGroup) {
		t.Fatalf("zero group = %v", err)
	}
	duplicate, _ := persons.NewGroup("work-colleagues")
	if _, err := persons.NewGroups(work, duplicate); !errors.Is(err, persons.ErrDuplicateGroup) {
		t.Fatalf("duplicate group = %v", err)
	}
	values := []persons.Group{work}
	groups, err := persons.NewGroups(values...)
	if err != nil {
		t.Fatal(err)
	}
	values[0] = persons.Group{}
	groups.Items()[0] = persons.Group{}
	groups.Strings()[0] = "changed"
	if !groups.Contains(work) || groups.Strings()[0] != "work-colleagues" {
		t.Fatal("groups changed through a caller-owned slice")
	}
}
