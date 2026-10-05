package habit

import (
	"context"
	"fmt"
	"time"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// CreateCommand defines an activity, its fixed goal and optional tags.
type CreateCommand struct {
	Content string
	Title   string
	Amount  float64
	Unit    string
	Period  string
	Tags    []string
}

// CreateResult contains the habit returned after a successful create operation.
type CreateResult struct {
	Habit *habits.Habit
}

// CreateHandler validates input, assigns identity and lifecycle metadata, and saves a habit.
type CreateHandler struct {
	repository ports.HabitRepository
	ids        ports.IDGenerator
}

// NewCreateHandler creates the use-case handler with its dependencies.
func NewCreateHandler(repository ports.HabitRepository, ids ports.IDGenerator) *CreateHandler {
	return &CreateHandler{
		repository: repository,
		ids:        ids,
	}
}

// Handle creates and persists a habit, returning it only after a successful save.
func (handler *CreateHandler) Handle(ctx context.Context, command CreateCommand) (CreateResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateResult{}, err
	}
	if handler.repository == nil || handler.ids == nil {
		return CreateResult{}, fmt.Errorf("create habit dependencies are not configured")
	}
	title, err := shared.NewTitle(command.Title)
	if err != nil {
		return CreateResult{}, err
	}
	unit, err := habits.NewUnit(command.Unit)
	if err != nil {
		return CreateResult{}, err
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
		return CreateResult{}, fmt.Errorf("unsupported habit period %q: use day, week or month", command.Period)
	}
	goal, err := habits.NewGoal(command.Amount, unit, period)
	if err != nil {
		return CreateResult{}, err
	}
	values := make([]shared.Tag, 0, len(command.Tags))
	for _, value := range command.Tags {
		tag, err := shared.NewTag(value)
		if err != nil {
			return CreateResult{}, err
		}
		values = append(values, tag)
	}
	tags, err := shared.NewTags(values...)
	if err != nil {
		return CreateResult{}, err
	}
	now := time.Now()
	metadata, err := shared.NewEntityMetadata(now, now)
	if err != nil {
		return CreateResult{}, err
	}
	habit, err := habits.NewHabit(handler.ids.Generate(), title, goal, tags, metadata)
	if err != nil {
		return CreateResult{}, err
	}
	if err := habit.Rewrite(command.Content, now); err != nil {
		return CreateResult{}, err
	}
	if err := handler.repository.Save(ctx, habit); err != nil {
		return CreateResult{}, fmt.Errorf("save habit: %w", err)
	}
	return CreateResult{Habit: habit}, nil
}
