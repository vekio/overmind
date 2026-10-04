package habits_test

import (
	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"math"
	"testing"
)

func TestGoalRequiresAmountUnitAndSupportedPeriod(t *testing.T) {
	amount := 2.0
	unit, _ := habits.NewUnit("litros")
	for _, input := range []struct {
		amount float64
		unit   habits.Unit
		period calendar.Period
	}{
		{0, unit, calendar.Day},
		{-1, unit, calendar.Day},
		{math.NaN(), unit, calendar.Day},
		{math.Inf(1), unit, calendar.Day},
		{math.Inf(-1), unit, calendar.Day},
		{amount, habits.Unit{}, calendar.Day},
		{amount, unit, calendar.Unknown},
		{amount, unit, calendar.Period(255)},
	} {
		if _, err := habits.NewGoal(input.amount, input.unit, input.period); err == nil {
			t.Fatal("invalid goal accepted")
		}
	}
	for _, period := range []calendar.Period{calendar.Day, calendar.Week, calendar.Month} {
		goal, err := habits.NewGoal(amount, unit, period)
		if err != nil {
			t.Fatal(err)
		}
		if goal.Amount() != amount || goal.Unit() != unit || goal.Period() != period {
			t.Fatal("goal lost its values")
		}
	}
}
