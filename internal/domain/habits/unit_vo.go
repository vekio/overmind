package habits

import (
	"fmt"
	"strings"
)

// Unit is a value object labeling what the target measures. Labels are literal; there is no unit conversion.
type Unit struct {
	value string
}

// NewUnit trims surrounding whitespace and requires a nonempty single-line label.
func NewUnit(value string) (Unit, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return Unit{}, fmt.Errorf("unit must be nonempty and single-line")
	}
	return Unit{value: value}, nil
}

// String returns the literal unit label.
func (unit Unit) String() string { return unit.value }

// IsZero reports whether the unit is uninitialized.
func (unit Unit) IsZero() bool { return unit.value == "" }
