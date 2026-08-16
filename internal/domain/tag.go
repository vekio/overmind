package domain

import (
	"errors"
	"fmt"
)

// ErrInvalidTag indicates that a tag has no usable characters after
// normalization.
var ErrInvalidTag = errors.New("invalid tag")

// Tag is the normalized value of one document tag.
type Tag struct {
	value string
}

// NewTag normalizes a human-readable tag using the domain slug policy.
func NewTag(raw string) (Tag, error) {
	value := Slugify(raw)
	if value == "" {
		return Tag{}, fmt.Errorf("%w: %q", ErrInvalidTag, raw)
	}

	return Tag{value: value}, nil
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
