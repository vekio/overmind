package app

import (
	"context"
	"fmt"
	"time"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// CreateHabitCommand defines an activity, its fixed goal and optional tags.
type CreateHabitCommand struct {
	Title  string
	Amount float64
	Unit   string
	Period string
	Tags   []string
}

type CreateHabitResult struct {
	Habit *habits.Habit
}

type CreateHabitHandler struct {
	repository ports.HabitRepository
	ids        ports.IDGenerator
}

func newCreateHabitHandler(repository ports.HabitRepository, ids ports.IDGenerator) *CreateHabitHandler {
	return &CreateHabitHandler{repository: repository, ids: ids}
}

// Handle creates and persists a habit, returning it only after a successful save.
func (handler *CreateHabitHandler) Handle(ctx context.Context, command CreateHabitCommand) (CreateHabitResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateHabitResult{}, err
	}
	if handler.repository == nil || handler.ids == nil {
		return CreateHabitResult{}, fmt.Errorf("create habit dependencies are not configured")
	}
	title, err := shared.NewTitle(command.Title)
	if err != nil {
		return CreateHabitResult{}, err
	}
	unit, err := habits.NewUnit(command.Unit)
	if err != nil {
		return CreateHabitResult{}, err
	}
	var period calendar.Period
	switch command.Period {
	case "day":
		period = calendar.Day
	case "week":
		period = calendar.Week
	case "month":
		period = calendar.Month
	default:
		return CreateHabitResult{}, fmt.Errorf("unsupported habit period %q: use day, week or month", command.Period)
	}
	goal, err := habits.NewGoal(command.Amount, unit, period)
	if err != nil {
		return CreateHabitResult{}, err
	}
	values := make([]shared.Tag, 0, len(command.Tags))
	for _, value := range command.Tags {
		tag, err := shared.NewTag(value)
		if err != nil {
			return CreateHabitResult{}, err
		}
		values = append(values, tag)
	}
	tags, err := shared.NewTags(values...)
	if err != nil {
		return CreateHabitResult{}, err
	}
	now := time.Now()
	metadata, err := shared.NewEntityMetadata(now, now)
	if err != nil {
		return CreateHabitResult{}, err
	}
	habit, err := habits.NewHabit(handler.ids.Generate(), title, goal, tags, metadata)
	if err != nil {
		return CreateHabitResult{}, err
	}
	if err := handler.repository.Save(ctx, habit); err != nil {
		return CreateHabitResult{}, fmt.Errorf("save habit: %w", err)
	}
	return CreateHabitResult{Habit: habit}, nil
}
