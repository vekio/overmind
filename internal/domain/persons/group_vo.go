package persons

import (
	"errors"
	"fmt"
	"github.com/vekio/overmind/internal/domain/shared"
)

// ErrInvalidGroup indicates an invalid group.
var ErrInvalidGroup = errors.New("invalid group")

// Group represents a normalized note group.
type Group struct {
	value string
}

// NewGroup creates a normalized group.
func NewGroup(value string) (Group, error) {
	normalized := shared.Slugify(value)
	if normalized == "" {
		return Group{}, fmt.Errorf("%w: %q", ErrInvalidGroup, value)
	}

	return Group{value: normalized}, nil
}

// String returns the normalized group value.
func (group Group) String() string {
	return group.value
}

// IsZero reports whether the group is uninitialized.
func (group Group) IsZero() bool {
	return group.value == ""
}

// Equal reports whether two groups have the same normalized value.
func (group Group) Equal(other Group) bool {
	return group.value == other.value
}
