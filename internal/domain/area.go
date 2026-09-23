package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// ErrInvalidArea indicates an invalid page area.
var ErrInvalidArea = errors.New("invalid area")

// Area represents a validated hierarchical location for pages.
type Area struct {
	segments []string
}

// NewArea creates an area from slash-separated segments.
func NewArea(value string) (Area, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Area{}, fmt.Errorf("%w: value is required", ErrInvalidArea)
	}

	parts := strings.Split(value, "/")
	segments := make([]string, 0, len(parts))
	for index, part := range parts {
		slug := Slugify(part)
		if slug == "" {
			return Area{}, fmt.Errorf("%w: segment %d must contain a letter or number", ErrInvalidArea, index+1)
		}

		segments = append(segments, slug)
	}

	return Area{segments: segments}, nil
}

// Name returns the first area segment.
func (area Area) Name() string {
	if area.IsZero() {
		return ""
	}

	return area.segments[0]
}

// IsZero reports whether the area is empty.
func (area Area) IsZero() bool {
	return len(area.segments) == 0
}

// Equal reports whether two areas contain the same segments.
func (area Area) Equal(other Area) bool {
	return slices.Equal(area.segments, other.segments)
}

// Subarea returns the next nested area when one exists.
func (area Area) Subarea() (Area, bool) {
	if len(area.segments) <= 1 {
		return Area{}, false
	}

	return Area{segments: slices.Clone(area.segments[1:])}, true
}

// String returns the normalized area path.
func (area Area) String() string {
	return strings.Join(area.segments, "/")
}
