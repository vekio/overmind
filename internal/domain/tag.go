package domain

import (
	"errors"
	"fmt"
)

// ErrInvalidTag indicates an invalid tag.
var ErrInvalidTag = errors.New("invalid tag")

// Tag represents a normalized note tag.
type Tag struct {
	value string
}

// NewTag creates a normalized tag.
func NewTag(value string) (Tag, error) {
	normalized := Slugify(value)
	if normalized == "" {
		return Tag{}, fmt.Errorf("%w: %q", ErrInvalidTag, value)
	}

	return Tag{value: normalized}, nil
}

// String returns the normalized tag value.
func (tag Tag) String() string {
	return tag.value
}

// IsZero reports whether the tag is uninitialized.
func (tag Tag) IsZero() bool {
	return tag.value == ""
}

// Equal reports whether two tags have the same normalized value.
func (tag Tag) Equal(other Tag) bool {
	return tag.value == other.value
}
