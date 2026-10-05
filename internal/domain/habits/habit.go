package habits

import (
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/shared"
)

// Habit is an entity identified by UUID. Its goal remains fixed so historical
// records keep their unit and target interpretation. A different goal requires
// a new habit until goal history is modeled explicitly.
type Habit struct {
	id       uuid.UUID
	title    shared.Title
	goal     Goal
	tags     shared.Tags
	content  string
	metadata shared.EntityMetadata
}

// NewHabit validates identity and required domain values without assigning timestamps.
func NewHabit(id uuid.UUID, title shared.Title, goal Goal, tags shared.Tags, metadata shared.EntityMetadata) (*Habit, error) {
	habit := &Habit{id: id, title: title, goal: goal, tags: tags, metadata: metadata}
	if err := habit.validate(); err != nil {
		return nil, err
	}
	return habit, nil
}

func (habit *Habit) validate() error {
	if habit == nil {
		return fmt.Errorf("habit is uninitialized")
	}
	if habit.id == uuid.Nil() {
		return fmt.Errorf("habit ID is required")
	}
	if habit.title.IsZero() {
		return fmt.Errorf("habit title is required")
	}
	if habit.metadata.IsZero() {
		return fmt.Errorf("habit metadata is required")
	}
	return habit.goal.validate()
}

// Metadata returns immutable creation and update timestamps.
func (habit *Habit) Metadata() shared.EntityMetadata { return habit.metadata }

// ID returns the stable document identity.
func (habit *Habit) ID() uuid.UUID { return habit.id }

// Title returns the validated activity name.
func (habit *Habit) Title() shared.Title { return habit.title }

// Goal returns the immutable target and its calendar period.
func (habit *Habit) Goal() Goal { return habit.goal }

// Tags returns the immutable ordered tag collection.
func (habit *Habit) Tags() shared.Tags { return habit.tags }

// Rename updates the title while preserving identity and goal.
func (habit *Habit) Rename(title shared.Title, at time.Time) error {
	if err := habit.validate(); err != nil {
		return err
	}
	if title.IsZero() {
		return fmt.Errorf("habit title is required")
	}
	metadata, err := habit.metadata.Updated(at)
	if err != nil {
		return err
	}
	habit.metadata = metadata
	habit.title = title
	return nil
}

// ReplaceTags replaces the collection; an empty collection clears it.
func (habit *Habit) ReplaceTags(tags shared.Tags, at time.Time) error {
	if err := habit.validate(); err != nil {
		return err
	}
	metadata, err := habit.metadata.Updated(at)
	if err != nil {
		return err
	}
	habit.metadata = metadata
	habit.tags = tags
	return nil
}

// Register creates an independent occurrence associated with this habit.
// The caller supplies the identifier; dates may be in the past or future.
func (habit *Habit) Register(id uuid.UUID, value float64, at time.Time) (*HabitRecord, error) {
	if err := habit.validate(); err != nil {
		return nil, err
	}
	return newHabitRecord(id, habit.id, value, at)
}

// Summary formats the activity name, target amount, unit and period.
func (habit *Habit) Summary() string {
	return fmt.Sprintf("%s: %g %s/%s", habit.title, habit.goal.Amount(), habit.goal.Unit(), habit.goal.Period())
}

// Content returns the original AsciiDoc body.
func (habit *Habit) Content() string { return habit.content }

// Rewrite replaces the body verbatim, including empty content.
// Invalid update times leave content and metadata unchanged.
func (habit *Habit) Rewrite(content string, at time.Time) error {
	if err := habit.validate(); err != nil {
		return err
	}
	metadata, err := habit.metadata.Updated(at)
	if err != nil {
		return err
	}
	habit.content, habit.metadata = content, metadata
	return nil
}
