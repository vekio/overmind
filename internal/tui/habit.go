package tui

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/habit"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/shared"
)

// CreateHabitFunc connects form submissions to the habit creation use case.
type CreateHabitFunc func(context.Context, habit.CreateCommand) (habit.CreateResult, error)

func newHabitForm() form {
	return newForm("Habit",
		fieldSpec{
			id: "title", label: "Title", kind: fieldText, placeholder: "Required, habit title",
			validate: func(raw string) error {
				_, err := shared.NewTitle(raw)
				return err
			},
		},
		fieldSpec{
			id: "amount", label: "Amount", kind: fieldText, placeholder: "Required, positive amount; example: 2.5",
			validate: func(raw string) error {
				_, err := habitAmount(raw)
				return err
			},
		},
		fieldSpec{
			id: "unit", label: "Unit", kind: fieldText, placeholder: "Required, example: litros or entrenamientos",
			validate: func(raw string) error {
				_, err := habits.NewUnit(raw)
				return err
			},
		},
		fieldSpec{
			id: "period", label: "Period", kind: fieldSelect, options: []string{"day", "week", "month"},
		},
		contentField("Write your habit…"),
		tagsField(),
	)
}

func habitAmount(raw string) (float64, error) {
	amount, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, fmt.Errorf("amount must be a finite number greater than zero")
	}
	return amount, nil
}

func (m model) saveHabit(values map[string]string) tea.Cmd {
	return func() tea.Msg {
		amount, err := habitAmount(values["amount"])
		if err != nil {
			return noteCreated{action: actionHabit, err: err}
		}
		if m.createHabit == nil {
			return noteCreated{action: actionHabit, err: fmt.Errorf("habit creation is not configured")}
		}

		command := habit.CreateCommand{
			Title:   values["title"],
			Content: values["content"],
			Amount:  amount,
			Unit:    values["unit"],
			Period:  values["period"],
			Tags:    formListValues(values["tags"]),
		}
		_, err = m.createHabit(m.ctx, command)
		return noteCreated{action: actionHabit, err: err}
	}
}
