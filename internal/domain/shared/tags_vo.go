package shared

import (
	"errors"
	"fmt"
	"slices"
)

// ErrDuplicateTag indicates a repeated normalized tag.
var ErrDuplicateTag = errors.New("duplicate tag")

// Tags is a value object: an immutable, ordered collection of unique tags.
// Its zero value is a valid empty collection.
type Tags struct {
	items []Tag
}

// NewTags creates an ordered collection of unique tags.
func NewTags(values ...Tag) (Tags, error) {
	seen := make(map[Tag]struct{}, len(values))
	for index, tag := range values {
		if tag.IsZero() {
			return Tags{}, fmt.Errorf("%w: tag at position %d is zero", ErrInvalidTag, index+1)
		}
		if _, exists := seen[tag]; exists {
			return Tags{}, fmt.Errorf("%w: %q", ErrDuplicateTag, tag)
		}
		seen[tag] = struct{}{}
	}

	return Tags{items: slices.Clone(values)}, nil
}

// Items returns a copy of the normalized tag values.
func (tags Tags) Items() []Tag {
	return slices.Clone(tags.items)
}

// Strings returns the normalized tags as strings in their original order.
func (tags Tags) Strings() []string {
	values := make([]string, len(tags.items))
	for index, tag := range tags.items {
		values[index] = tag.String()
	}

	return values
}

// Len returns the number of tags.
func (tags Tags) Len() int {
	return len(tags.items)
}

// IsEmpty reports whether the collection contains no tags.
func (tags Tags) IsEmpty() bool {
	return len(tags.items) == 0
}

// Contains reports whether the collection contains tag.
func (tags Tags) Contains(tag Tag) bool {
	for _, existing := range tags.items {
		if existing.Equal(tag) {
			return true
		}
	}

	return false
}
