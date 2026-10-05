package person_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	appperson "github.com/vekio/overmind/internal/app/person"
	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/ports"
)

type personTestRepository struct {
	person *persons.Person
	err    error
}

func (repo *personTestRepository) Save(_ context.Context, person *persons.Person) error {
	if repo.err != nil {
		return repo.err
	}
	repo.person = person
	return nil
}
func (repo *personTestRepository) ByID(context.Context, uuid.UUID) (*persons.Person, error) {
	if repo.person == nil {
		return nil, ports.ErrPersonNotFound
	}
	return repo.person, nil
}

func TestCreatePerson(t *testing.T) {
	repo := &personTestRepository{}
	application := app.New(app.Dependencies{Persons: repo, IDGenerator: testIDGenerator{}})
	result, err := application.Commands.CreatePerson.Handle(context.Background(), appperson.CreateCommand{Name: " Ana García ", Groups: []string{"Trabajo", "Universidad"}, Tags: []string{"Amiga"}})
	if err != nil {
		t.Fatal(err)
	}
	person := result.Person
	if person != repo.person || person.ID() == uuid.Nil() || person.Name().String() != "Ana García" || !reflect.DeepEqual(person.Groups().Strings(), []string{"trabajo", "universidad"}) || !reflect.DeepEqual(person.Tags().Strings(), []string{"amiga"}) {
		t.Fatal("raw command did not create and persist validated domain values")
	}
	if person.Metadata().IsZero() || !person.Metadata().CreatedAt().Equal(person.Metadata().UpdatedAt()) {
		t.Fatal("initial timestamps must match")
	}
}
func TestCreatePersonRejectsInvalidInputWithoutSaving(t *testing.T) {
	for _, command := range []appperson.CreateCommand{
		{}, {Name: "Ana", Groups: []string{"!!!"}}, {Name: "Ana", Groups: []string{"Work", "work"}}, {Name: "Ana", Tags: []string{"!!!"}}, {Name: "Ana", Tags: []string{"Amiga", "amiga"}},
	} {
		repo := &personTestRepository{}
		application := app.New(app.Dependencies{Persons: repo, IDGenerator: testIDGenerator{}})
		result, err := application.Commands.CreatePerson.Handle(context.Background(), command)
		if err == nil || result.Person != nil || repo.person != nil {
			t.Fatal("invalid command saved a person")
		}
	}
}
func TestCreatePersonSupportsEmptyGroupsAndTagsAndPropagatesFailures(t *testing.T) {
	ctx := context.Background()
	repo := &personTestRepository{}
	application := app.New(app.Dependencies{Persons: repo, IDGenerator: testIDGenerator{}})
	command := appperson.CreateCommand{Name: "Ana"}
	created, err := application.Commands.CreatePerson.Handle(ctx, command)
	if err != nil || !created.Person.Groups().IsEmpty() || !created.Person.Tags().IsEmpty() {
		t.Fatalf("optional groups/tags = %v", err)
	}
	failure := errors.New("disk failure")
	repo.err = failure
	if result, err := application.Commands.CreatePerson.Handle(ctx, command); !errors.Is(err, failure) || result.Person != nil {
		t.Fatalf("save failure = %v", err)
	}
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := application.Commands.CreatePerson.Handle(ctx, command); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := app.New(app.Dependencies{}).Commands.CreatePerson.Handle(context.Background(), command); err == nil {
		t.Fatal("missing dependencies accepted")
	}
}

// testIDGenerator keeps ID generation independent of infrastructure.
type testIDGenerator struct{}

func (testIDGenerator) Generate() uuid.UUID { return uuid.New() }

func (repo *personTestRepository) Update(ctx context.Context, entity *persons.Person) error {
	return repo.Save(ctx, entity)
}
