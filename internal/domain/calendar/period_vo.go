package calendar

import (
	"fmt"
	"time"
)

// Period is a value object identifying how often a habit's target applies.
type Period uint8

const (
	Unknown Period = iota
	Day
	Week
	Month
)

// IsValid reports whether the period is supported.
func (period Period) IsValid() bool {
	switch period {
	case Day, Week, Month:
		return true
	default:
		return false
	}
}

func (period Period) String() string {
	switch period {
	case Day:
		return "day"
	case Week:
		return "week"
	case Month:
		return "month"
	default:
		return "unknown"
	}
}

// Window returns the calendar interval containing date. Weeks start on Monday;
// months span from their first day to the first day of the next month.
func (period Period) Window(date Date) (Window, error) {
	if !period.IsValid() {
		return Window{}, fmt.Errorf("unsupported period %q", period)
	}
	if date.IsZero() {
		return Window{}, fmt.Errorf("date is required")
	}
	start := date.value
	var end time.Time
	switch period {
	case Day:
		end = start.AddDate(0, 0, 1)
	case Week:
		daysSinceMonday := (int(start.Weekday()) + 6) % 7
		start = start.AddDate(0, 0, -daysSinceMonday)
		end = start.AddDate(0, 0, 7)
	case Month:
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0)
	}
	return NewWindow(Date{value: start, initialized: true}, Date{value: end, initialized: true})
}
