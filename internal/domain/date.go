package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInvalidDate indicates an invalid calendar date.
var ErrInvalidDate = errors.New("invalid date")

// Date represents a calendar date in YYYY-MM-DD format.
type Date struct {
	value string
}

// NewDate creates a validated calendar date.
func NewDate(value string) (Date, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Date{}, fmt.Errorf("%w: value is required", ErrInvalidDate)
	}
	if _, err := time.Parse(time.DateOnly, value); err != nil {
		return Date{}, fmt.Errorf("%w: %q must use YYYY-MM-DD format", ErrInvalidDate, value)
	}

	return Date{value: value}, nil
}

// DateFromTime creates a date from a time value.
func DateFromTime(value time.Time) Date {
	if value.IsZero() {
		return Date{}
	}

	return Date{value: value.Format(time.DateOnly)}
}

// IsZero reports whether the date is uninitialized.
func (date Date) IsZero() bool {
	return date.value == ""
}

// String returns the date in YYYY-MM-DD format.
func (date Date) String() string {
	return date.value
}
