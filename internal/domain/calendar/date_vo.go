package calendar

import (
	"errors"
	"fmt"
	"time"
)

var ErrInvalidDate = errors.New("invalid date")

// Date is a value object representing a calendar date without a time or timezone. Internally it uses UTC
// midnight for comparison and date arithmetic; callers choose the local date.
type Date struct {
	value       time.Time
	initialized bool
}

func NewDate(value string) (Date, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return Date{}, fmt.Errorf("%w %q: %w", ErrInvalidDate, value, err)
	}
	return Date{value: parsed, initialized: true}, nil
}

func (date Date) String() string { return date.value.Format(time.DateOnly) }

// IsZero reports whether the date is uninitialized. The initialization flag
// distinguishes Date{} from the valid calendar date 0001-01-01.
func (date Date) IsZero() bool { return !date.initialized }

// Before reports whether date precedes other. Both dates must be initialized.
func (date Date) Before(other Date) bool {
	return !date.IsZero() && !other.IsZero() && date.value.Before(other.value)
}

// Equal reports whether both dates represent the same date or are both zero.
func (date Date) Equal(other Date) bool {
	return date.initialized == other.initialized && date.value.Equal(other.value)
}

// DateFromTime uses the caller's local date rather than converting to UTC first.
func DateFromTime(value time.Time) Date {
	if value.IsZero() {
		return Date{}
	}
	date, _ := NewDate(value.Format(time.DateOnly))
	return date
}
