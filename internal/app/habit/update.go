package habit

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// UpdateCommand replaces an activity, goal, tags and descriptive content.
// The use case reconstructs the entity so it can replace the domain entity's fixed goal.
type UpdateCommand struct {
	Content string
	ID      string
	Title   string
	Amount  float64
	Unit    string
	Period  string
	Tags    []string
}

// UpdateResult contains the habit returned after a successful update operation.
type UpdateResult struct {
	Habit *habits.Habit
}

// UpdateHandler validates replacement values and retains the habit's identity and creation time.
type UpdateHandler struct {
	repository ports.HabitRepository
}

// NewUpdateHandler creates the use-case handler with its dependencies.
func NewUpdateHandler(repository ports.HabitRepository) *UpdateHandler {
	return &UpdateHandler{
		repository: repository,
	}
}

// Handle validates and persists changes to an existing habit.
func (handler *UpdateHandler) Handle(ctx context.Context, command UpdateCommand) (UpdateResult, error) {
	if err := ctx.Err(); err != nil {
		return UpdateResult{}, err
	}
	if handler.repository == nil {
		return UpdateResult{}, fmt.Errorf("update habit dependencies are not configured")
	}
	title, err := shared.NewTitle(command.Title)
	if err != nil {
		return UpdateResult{}, err
	}
	unit, err := habits.NewUnit(command.Unit)
	if err != nil {
		return UpdateResult{}, err
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
		return UpdateResult{}, fmt.Errorf("unsupported habit period %q: use day, week or month", command.Period)
	}
	goal, err := habits.NewGoal(command.Amount, unit, period)
	if err != nil {
		return UpdateResult{}, err
	}
	values := make([]shared.Tag, 0, len(command.Tags))
	for _, value := range command.Tags {
		tag, err := shared.NewTag(value)
		if err != nil {
			return UpdateResult{}, err
		}
		values = append(values, tag)
	}
	tags, err := shared.NewTags(values...)
	if err != nil {
		return UpdateResult{}, err
	}
	id, err := uuid.Parse(command.ID)
	if err != nil || id == uuid.Nil() {
		return UpdateResult{}, fmt.Errorf("valid habit ID is required")
	}
	existing, err := handler.repository.ByID(ctx, id)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("read habit: %w", err)
	}
	if existing == nil || existing.ID() != id {
		return UpdateResult{}, fmt.Errorf("habit identity does not match requested ID")
	}
	now := time.Now()
	metadata, err := existing.Metadata().Updated(now)
	if err != nil {
		return UpdateResult{}, err
	}
	habit, err := habits.NewHabit(id, title, goal, tags, metadata)
	if err != nil {
		return UpdateResult{}, err
	}
	if err := habit.Rewrite(command.Content, now); err != nil {
		return UpdateResult{}, err
	}
	if err := handler.repository.Update(ctx, habit); err != nil {
		return UpdateResult{}, fmt.Errorf("update habit: %w", err)
	}
	return UpdateResult{Habit: habit}, nil
}
