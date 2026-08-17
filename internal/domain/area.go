package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var ErrInvalidArea = errors.New("invalid area")

// Area represents a validated hierarchical location for pages.
type Area struct {
	segments []string
}

// NewArea creates a normalized Area from slash-separated path segments.
func NewArea(raw string) (Area, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return Area{}, fmt.Errorf("%w: value is required", ErrInvalidArea)
	}

	parts := strings.Split(v, "/")
	return buildArea(parts)
}

func buildArea(parts []string) (Area, error) {
	if len(parts) == 0 {
		return Area{}, fmt.Errorf("%w: value is required", ErrInvalidArea)
	}

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

func (a Area) Name() string {
	if a.IsZero() {
		return ""
	}

	return a.segments[0]
}

func (a Area) IsZero() bool {
	return len(a.segments) == 0
}

func (a Area) Equal(other Area) bool {
	return slices.Equal(a.segments, other.segments)
}

// Subarea returns the next nested area when one exists.
func (a Area) Subarea() (Area, bool) {
	if len(a.segments) <= 1 {
		return Area{}, false
	}

	return Area{segments: append([]string(nil), a.segments[1:]...)}, true
}

func (a Area) String() string {
	return strings.Join(a.segments, "/")
}
