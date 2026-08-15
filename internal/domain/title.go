package domain

import (
	"errors"
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
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return Title{}, ErrInvalidTitle
	}
	slug := Slugify(value)
	if slug == "" {
		return Title{}, ErrInvalidTitle
	}

	return Title{value: value, slug: slug}, nil
}

// IsZero reports whether title is the invalid zero value.
func (title Title) IsZero() bool { return title.value == "" }

// String returns the title text.
func (title Title) String() string { return title.value }

// Slug returns the normalized slug generated from the title.
func (title Title) Slug() string { return title.slug }
