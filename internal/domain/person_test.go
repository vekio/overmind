package domain_test

import (
	"errors"
	"testing"
	"time"
	"uuid"

	"git.casta.me/alberto/overmind/internal/domain"
)

func TestPersonRequiresNameAndMetadata(t *testing.T) {
	name, _ := domain.NewTitle("Ana García")
	for _, input := range []struct {
		id   uuid.UUID
		name domain.Title
		when time.Time
	}{
		{uuid.New(), domain.Title{}, time.Now()},
		{uuid.Nil(), name, time.Now()},
		{uuid.New(), name, time.Time{}},
	} {
		if _, err := domain.NewPerson(input.id, input.name, domain.Groups{}, domain.Tags{}, input.when); !errors.Is(err, domain.ErrInvalidPerson) {
			t.Fatalf("invalid person = %v", err)
		}
	}
}

func TestGroupsAreNormalizedUniqueAndImmutable(t *testing.T) {
	work, err := domain.NewGroup(" Work colleagues ")
	if err != nil || work.String() != "work-colleagues" {
		t.Fatalf("normalized group = %q, %v", work, err)
	}
	if _, err := domain.NewGroup(" !!! "); !errors.Is(err, domain.ErrInvalidGroup) {
		t.Fatalf("invalid group = %v", err)
	}
	if _, err := domain.NewGroups(domain.Group{}); !errors.Is(err, domain.ErrInvalidGroup) {
		t.Fatalf("zero group = %v", err)
	}
	duplicate, _ := domain.NewGroup("work-colleagues")
	if _, err := domain.NewGroups(work, duplicate); !errors.Is(err, domain.ErrDuplicateGroup) {
		t.Fatalf("duplicate group = %v", err)
	}
	values := []domain.Group{work}
	groups, err := domain.NewGroups(values...)
	if err != nil {
		t.Fatal(err)
	}
	values[0] = domain.Group{}
	groups.Items()[0] = domain.Group{}
	groups.Strings()[0] = "changed"
	if !groups.Contains(work) || groups.Strings()[0] != "work-colleagues" {
		t.Fatal("groups changed through a caller-owned slice")
	}
}
