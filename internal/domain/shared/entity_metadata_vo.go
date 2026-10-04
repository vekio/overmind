package shared

import (
	"errors"
	"fmt"
	"time"
)

var ErrInvalidEntityMetadata = errors.New("invalid entity metadata")

// EntityMetadata is an immutable value object describing an entity's lifetime.
type EntityMetadata struct {
	createdAt time.Time
	updatedAt time.Time
}

func NewEntityMetadata(createdAt, updatedAt time.Time) (EntityMetadata, error) {
	if createdAt.IsZero() || updatedAt.IsZero() || updatedAt.Before(createdAt) {
		return EntityMetadata{}, fmt.Errorf("%w: nonzero times with updatedAt >= createdAt are required", ErrInvalidEntityMetadata)
	}
	return EntityMetadata{createdAt: createdAt.UTC().Round(0), updatedAt: updatedAt.UTC().Round(0)}, nil
}
func (metadata EntityMetadata) CreatedAt() time.Time { return metadata.createdAt }
func (metadata EntityMetadata) UpdatedAt() time.Time { return metadata.updatedAt }
func (metadata EntityMetadata) IsZero() bool         { return metadata.createdAt.IsZero() }

// Updated returns a copy with a later or equal update time.
func (metadata EntityMetadata) Updated(at time.Time) (EntityMetadata, error) {
	if metadata.IsZero() || at.IsZero() || at.Before(metadata.updatedAt) {
		return EntityMetadata{}, fmt.Errorf("%w: update time must not precede the previous update", ErrInvalidEntityMetadata)
	}
	return NewEntityMetadata(metadata.createdAt, at)
}
