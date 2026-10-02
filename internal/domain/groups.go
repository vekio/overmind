package domain

import (
	"errors"
	"fmt"
	"slices"
)

// ErrDuplicateGroup indicates a repeated normalized group.
var ErrDuplicateGroup = errors.New("duplicate group")

// Groups is an immutable, ordered collection of unique note groups.
// Its zero value is a valid empty collection.
type Groups struct {
	items []Group
}

// NewGroups creates an ordered collection of unique groups.
func NewGroups(values ...Group) (Groups, error) {
	seen := make(map[Group]struct{}, len(values))
	for index, group := range values {
		if group.IsZero() {
			return Groups{}, fmt.Errorf("%w: group at position %d is zero", ErrInvalidGroup, index+1)
		}
		if _, exists := seen[group]; exists {
			return Groups{}, fmt.Errorf("%w: %q", ErrDuplicateGroup, group)
		}
		seen[group] = struct{}{}
	}

	return Groups{items: slices.Clone(values)}, nil
}

// Items returns a copy of the normalized group values.
func (groups Groups) Items() []Group {
	return slices.Clone(groups.items)
}

// Strings returns the normalized groups as strings in their original order.
func (groups Groups) Strings() []string {
	values := make([]string, len(groups.items))
	for index, group := range groups.items {
		values[index] = group.String()
	}

	return values
}

// Len returns the number of groups.
func (groups Groups) Len() int {
	return len(groups.items)
}

// IsEmpty reports whether the collection contains no groups.
func (groups Groups) IsEmpty() bool {
	return len(groups.items) == 0
}

// Contains reports whether the collection contains group.
func (groups Groups) Contains(group Group) bool {
	for _, existing := range groups.items {
		if existing.Equal(group) {
			return true
		}
	}

	return false
}
