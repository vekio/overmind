package habits_test

import (
	"github.com/vekio/overmind/internal/domain/habits"
	"math"
	"testing"
	"time"
	"uuid"
)

func TestOccurrencesAreIndependent(t *testing.T) {
	habit := waterHabit(t)
	at := time.Date(2026, 10, 3, 13, 0, 0, 0, time.FixedZone("local", 2*60*60))
	first, err := habit.Register(uuid.New(), 1, at)
	if err != nil {
		t.Fatal(err)
	}
	second, err := habit.Register(uuid.New(), 1, at)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID() == second.ID() || first.HabitID() != habit.ID() || second.HabitID() != habit.ID() || first.Value() != 1 || second.Value() != 1 || !first.OccurredAt().Equal(at) || !second.OccurredAt().Equal(at) {
		t.Fatal("same-instant occurrences must retain independent identities and their values")
	}
}

func TestRegisterRejectsInvalidInputs(t *testing.T) {
	habit := waterHabit(t)
	for _, input := range []struct {
		id    uuid.UUID
		value float64
		at    time.Time
	}{
		{uuid.Nil(), 1, time.Now()},
		{uuid.New(), -1, time.Now()},
		{uuid.New(), math.NaN(), time.Now()},
		{uuid.New(), math.Inf(1), time.Now()},
		{uuid.New(), math.Inf(-1), time.Now()},
		{uuid.New(), 1, time.Time{}},
	} {
		if record, err := habit.Register(input.id, input.value, input.at); err == nil || record != nil {
			t.Fatal("invalid occurrence accepted")
		}
	}
	for _, invalidHabit := range []*habits.Habit{nil, {}} {
		if record, err := invalidHabit.Register(uuid.New(), 1, time.Now()); err == nil || record != nil {
			t.Fatal("uninitialized habit registered an occurrence")
		}
	}
}

func TestRegisterAllowsZeroPositiveValuesAndPastFutureTimes(t *testing.T) {
	habit := waterHabit(t)
	now := time.Now()
	for _, at := range []time.Time{now.AddDate(-10, 0, 0), now.AddDate(10, 0, 0)} {
		for _, value := range []float64{0, 0.5, 1, 4} {
			record, err := habit.Register(uuid.New(), value, at)
			if err != nil {
				t.Fatal(err)
			}
			if record.Value() != value || !record.OccurredAt().Equal(at) {
				t.Fatal("occurrence lost its value or timestamp")
			}
		}
	}
}
