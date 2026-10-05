package habits

import (
	"fmt"
	"math"

	"github.com/vekio/overmind/internal/domain/calendar"
)

// Goal is an immutable value object describing an amount and unit per period.
type Goal struct {
	amount float64
	unit   Unit
	period calendar.Period
}

// NewGoal requires a finite positive amount, an initialized unit and a supported period.
func NewGoal(amount float64, unit Unit, period calendar.Period) (Goal, error) {
	goal := Goal{amount: amount, unit: unit, period: period}
	if err := goal.validate(); err != nil {
		return Goal{}, err
	}
	return goal, nil
}

func (goal Goal) validate() error {
	if goal.amount <= 0 || math.IsNaN(goal.amount) || math.IsInf(goal.amount, 0) {
		return fmt.Errorf("goal amount must be finite and greater than zero")
	}
	if goal.unit.IsZero() {
		return fmt.Errorf("goal unit is required")
	}
	if !goal.period.IsValid() {
		return fmt.Errorf("unsupported goal period %q", goal.period)
	}
	return nil
}

// Amount returns the required quantity per period.
func (goal Goal) Amount() float64 { return goal.amount }

// Unit returns the literal measurement label.
func (goal Goal) Unit() Unit { return goal.unit }

// Period returns the calendar period to which the target applies.
func (goal Goal) Period() calendar.Period { return goal.period }
