package domain

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidTitle indicates an invalid note title.
var ErrInvalidTitle = errors.New("invalid title")

// Title represents a validated, single-line note title.
type Title struct {
	value string
	slug  string
}

// NewTitle creates a title and its normalized slug.
func NewTitle(value string) (Title, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Title{}, fmt.Errorf("%w: value is required", ErrInvalidTitle)
	}
	if strings.ContainsAny(value, "\r\n") {
		return Title{}, fmt.Errorf("%w: value must be a single line", ErrInvalidTitle)
	}
	slug := Slugify(value)
	if slug == "" {
		return Title{}, fmt.Errorf("%w: value must contain a letter or number", ErrInvalidTitle)
	}

	return Title{value: value, slug: slug}, nil
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
	return title.slug
}
