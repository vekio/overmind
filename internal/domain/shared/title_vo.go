package shared

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidTitle indicates an invalid title.
var ErrInvalidTitle = errors.New("invalid title")

// Title represents a validated, single-line title.
type Title struct {
	value string
}

// NewTitle creates a validated title that can produce a nonempty slug.
func NewTitle(value string) (Title, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Title{}, fmt.Errorf("%w: value is required", ErrInvalidTitle)
	}
	if strings.ContainsAny(value, "\r\n") {
		return Title{}, fmt.Errorf("%w: value must be a single line", ErrInvalidTitle)
	}
	if Slugify(value) == "" {
		return Title{}, fmt.Errorf("%w: value must contain a letter or number", ErrInvalidTitle)
	}

	return Title{value: value}, nil
}

// IsZero reports whether the title is uninitialized.
func (title Title) IsZero() bool {
	return title.value == ""
}

// String returns the title text.
func (title Title) String() string {
	return title.value
}

// Slug returns the normalized title slug.
func (title Title) Slug() string {
	return Slugify(title.value)
}
