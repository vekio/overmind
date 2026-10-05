package habit_test

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	apphabit "github.com/vekio/overmind/internal/app/habit"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/ports"
)

func waterCommand() apphabit.CreateCommand {
	return apphabit.CreateCommand{Title: "Beber agua", Amount: 2, Unit: "litros", Period: "day", Tags: []string{"Bienestar"}}
}

func TestCreateHabit(t *testing.T) {
	store := &habitTestStore{}
	result, err := habitApplication(store).Commands.CreateHabit.Handle(context.Background(), waterCommand())
	if err != nil {
		t.Fatal(err)
	}
	habit := result.Habit
	if habit.Metadata().IsZero() || !habit.Metadata().CreatedAt().Equal(habit.Metadata().UpdatedAt()) {
		t.Fatal("new habit must share creation and initial modification time")
	}
	if habit != store.habit || habit.ID() == uuid.Nil() {
		t.Fatal("created habit was not persisted")
	}
	if habit.Summary() != "Beber agua: 2 litros/day" || !reflect.DeepEqual(habit.Tags().Strings(), []string{"bienestar"}) {
		t.Fatal("input was not converted into the expected domain values")
	}
}

func TestCreateHabitRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name   string
		change func(*apphabit.CreateCommand)
	}{
		{"empty title", func(c *apphabit.CreateCommand) { c.Title = "" }},
		{"empty unit", func(c *apphabit.CreateCommand) { c.Unit = "" }},
		{"zero amount", func(c *apphabit.CreateCommand) { c.Amount = 0 }},
		{"negative amount", func(c *apphabit.CreateCommand) { c.Amount = -1 }},
		{"nonfinite amount", func(c *apphabit.CreateCommand) { c.Amount = math.NaN() }},
		{"invalid period", func(c *apphabit.CreateCommand) { c.Period = "year" }},
		{"invalid tag", func(c *apphabit.CreateCommand) { c.Tags = []string{"---"} }},
		{"duplicate tags", func(c *apphabit.CreateCommand) { c.Tags = []string{"Bienestar", "bienestar"} }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			store := &habitTestStore{}
			command := waterCommand()
			test.change(&command)
			result, err := habitApplication(store).Commands.CreateHabit.Handle(context.Background(), command)
			if err == nil || result.Habit != nil || store.habit != nil {
				t.Fatal("invalid input must not be persisted")
			}
		})
	}
}

func TestCreateHabitSupportsPeriodsAndOptionalTags(t *testing.T) {
	for _, period := range []string{"day", "week", "month"} {
		t.Run(period, func(t *testing.T) {
			command := waterCommand()
			command.Period, command.Tags = period, nil
			result, err := habitApplication(&habitTestStore{}).Commands.CreateHabit.Handle(context.Background(), command)
			if err != nil {
				t.Fatal(err)
			}
			if result.Habit.Goal().Period().String() != period || !result.Habit.Tags().IsEmpty() {
				t.Fatal("period or optional tags were lost")
			}
		})
	}
}

func TestCreateHabitPropagatesSaveFailure(t *testing.T) {
	failure := errors.New("storage unavailable")
	result, err := habitApplication(&habitTestStore{saveErr: failure}).Commands.CreateHabit.Handle(context.Background(), waterCommand())
	if !errors.Is(err, failure) || result.Habit != nil {
		t.Fatalf("save failure = %v", err)
	}
}

func TestCreateHabitHonorsCancellationAndMissingDependencies(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := &habitTestStore{}
	if _, err := habitApplication(store).Commands.CreateHabit.Handle(ctx, waterCommand()); !errors.Is(err, context.Canceled) || store.habit != nil {
		t.Fatalf("canceled creation = %v", err)
	}
	if _, err := app.New(app.Dependencies{}).Commands.CreateHabit.Handle(context.Background(), waterCommand()); err == nil {
		t.Fatal("missing dependencies accepted")
	}
}

type habitTestStore struct {
	habit   *habits.Habit
	saveErr error
}

func (store *habitTestStore) Save(_ context.Context, habit *habits.Habit) error {
	if store.saveErr != nil {
		return store.saveErr
	}
	store.habit = habit
	return nil
}

func (store *habitTestStore) ByID(_ context.Context, id uuid.UUID) (*habits.Habit, error) {
	if store.habit == nil || store.habit.ID() != id {
		return nil, ports.ErrHabitNotFound
	}
	return store.habit, nil
}

func habitApplication(store *habitTestStore) *app.Application {
	return app.New(app.Dependencies{IDGenerator: testIDGenerator{}, Habits: store})
}

// testIDGenerator keeps ID generation independent of infrastructure.
type testIDGenerator struct{}

func (testIDGenerator) Generate() uuid.UUID { return uuid.New() }

func (store *habitTestStore) Update(ctx context.Context, entity *habits.Habit) error {
	return store.Save(ctx, entity)
}
