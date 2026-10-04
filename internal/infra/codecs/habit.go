package codecs

import (
	"fmt"
	"strconv"
	"uuid"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

func (HabitCodec) Encode(habit *habits.Habit) ([]byte, error) {
	if habit == nil || habit.ID() == uuid.Nil() {
		return nil, fmt.Errorf("initialized habit is required")
	}
	return render("habit", habit)
}

func (HabitCodec) Decode(source []byte) (*habits.Habit, error) {
	note, err := decodeHeader(source, ports.NoteKindHabit.String())
	if err != nil {
		return nil, err
	}
	header := note.header
	if header.Title == nil {
		return nil, fmt.Errorf("habit title is required")
	}
	title, err := shared.NewTitle(header.Title.Text)
	if err != nil {
		return nil, err
	}
	amountText, err := required(header.Attributes, "overmind-amount")
	if err != nil {
		return nil, err
	}
	amount, err := strconv.ParseFloat(amountText, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid habit amount: %w", err)
	}
	unitText, err := required(header.Attributes, "overmind-unit")
	if err != nil {
		return nil, err
	}
	unit, err := habits.NewUnit(unitText)
	if err != nil {
		return nil, err
	}
	periodText, err := required(header.Attributes, "overmind-period")
	if err != nil {
		return nil, err
	}
	var period calendar.Period
	switch periodText {
	case "day":
		period = calendar.Day
	case "week":
		period = calendar.Week
	case "month":
		period = calendar.Month
	default:
		return nil, fmt.Errorf("unsupported habit period %q", periodText)
	}
	goal, err := habits.NewGoal(amount, unit, period)
	if err != nil {
		return nil, err
	}
	return habits.NewHabit(note.id, title, goal, note.tags, note.metadata)
}

type HabitCodec struct{}

var _ ports.HabitCodec = HabitCodec{}
