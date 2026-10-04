package habits_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/shared"
)

func waterHabit(t *testing.T) *habits.Habit {
	t.Helper()
	title, err := shared.NewTitle("Beber agua")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := habits.NewUnit("litros")
	if err != nil {
		t.Fatal(err)
	}
	goal, err := habits.NewGoal(2, unit, calendar.Day)
	if err != nil {
		t.Fatal(err)
	}
	habit, err := habits.NewHabit(uuid.New(), title, goal, shared.Tags{}, fixtureMetadata())
	if err != nil {
		t.Fatal(err)
	}
	return habit
}

func ExampleNewHabit() {
	title, _ := shared.NewTitle("Beber agua")
	unit, _ := habits.NewUnit("litros")
	goal, _ := habits.NewGoal(2, unit, calendar.Day)
	habit, _ := habits.NewHabit(uuid.New(), title, goal, shared.Tags{}, fixtureMetadata())
	fmt.Println(habit.Summary())
	// Output: Beber agua: 2 litros/day
}

func TestHabitRequiresIdentityTitleAndGoal(t *testing.T) {
	valid := waterHabit(t)
	for _, input := range []struct {
		id    uuid.UUID
		title shared.Title
		goal  habits.Goal
	}{
		{uuid.Nil(), valid.Title(), valid.Goal()},
		{valid.ID(), shared.Title{}, valid.Goal()},
		{valid.ID(), valid.Title(), habits.Goal{}},
	} {
		if habit, err := habits.NewHabit(input.id, input.title, input.goal, shared.Tags{}, fixtureMetadata()); err == nil || habit != nil {
			t.Fatal("invalid habit accepted")
		}
	}
}

func TestHabitEditsPreserveIdentityAndGoal(t *testing.T) {
	habit := waterHabit(t)
	alias := habit
	id, goal := habit.ID(), habit.Goal()
	title, _ := shared.NewTitle("Hidratarse")
	tag, _ := shared.NewTag("Bienestar")
	tags, _ := shared.NewTags(tag)
	if err := habit.Rename(title, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := habit.ReplaceTags(tags, time.Now()); err != nil {
		t.Fatal(err)
	}
	if alias.Title() != title || !alias.Tags().Contains(tag) || habit.ID() != id || habit.Goal() != goal {
		t.Fatal("edits must preserve entity identity and goal")
	}
	items := habit.Tags().Items()
	items[0] = shared.Tag{}
	if !habit.Tags().Contains(tag) {
		t.Fatal("tags changed through returned slice")
	}
	before := *habit
	if err := habit.Rename(shared.Title{}, time.Now()); err == nil || !reflect.DeepEqual(*habit, before) {
		t.Fatal("invalid rename must preserve the entity")
	}
	if err := habit.ReplaceTags(shared.Tags{}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if !alias.Tags().IsEmpty() || habit.ID() != id || habit.Goal() != goal {
		t.Fatal("clearing tags changed identity or goal")
	}
}

func TestUninitializedHabitsCannotBeEdited(t *testing.T) {
	title, _ := shared.NewTitle("Agua")
	for _, habit := range []*habits.Habit{nil, {}} {
		if err := habit.Rename(title, time.Now()); err == nil {
			t.Fatal("uninitialized habit renamed")
		}
		if err := habit.ReplaceTags(shared.Tags{}, time.Now()); err == nil {
			t.Fatal("uninitialized habit edited")
		}
	}
}

func fixtureMetadata() shared.EntityMetadata {
	at := time.Date(2026, 10, 3, 12, 0, 0, 123456789, time.UTC)
	metadata, err := shared.NewEntityMetadata(at, at)
	if err != nil {
		panic(err)
	}
	return metadata
}

func TestHabitChangesUpdateMetadataAtomically(t *testing.T) {
	habit := waterHabit(t)
	created := habit.Metadata().CreatedAt()
	renamed, _ := shared.NewTitle("Hidratarse")
	at := created.Add(time.Hour)
	if err := habit.Rename(renamed, at); err != nil {
		t.Fatal(err)
	}
	if !habit.Metadata().CreatedAt().Equal(created) || !habit.Metadata().UpdatedAt().Equal(at) {
		t.Fatal("rename must update only the modification time")
	}
	before := *habit
	if err := habit.ReplaceTags(shared.Tags{}, created); err == nil || !reflect.DeepEqual(*habit, before) {
		t.Fatal("backdated update must not partially mutate habit")
	}
	if err := habit.Rename(renamed, time.Time{}); err == nil || !reflect.DeepEqual(*habit, before) {
		t.Fatal("zero update time accepted")
	}
	if _, err := habits.NewHabit(habit.ID(), habit.Title(), habit.Goal(), habit.Tags(), shared.EntityMetadata{}); err == nil {
		t.Fatal("uninitialized metadata accepted")
	}
}
