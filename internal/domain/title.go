package domain

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidTitle = errors.New("invalid title")

// Title is a validated, single-line document title.
type Title struct {
	value string
	slug  string
}

// NewTitle creates a normalized Title.
func NewTitle(raw string) (Title, error) {
	value := strings.TrimSpace(raw)
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

// IsZero reports whether title is the invalid zero value.
func (title Title) IsZero() bool { return title.value == "" }

// String returns the title text.
func (title Title) String() string { return title.value }

// Slug returns the normalized slug generated from the title.
func (title Title) Slug() string { return title.slug }
