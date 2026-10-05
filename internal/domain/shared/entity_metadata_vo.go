package shared

import (
	"errors"
	"fmt"
	"time"
)

// ErrInvalidEntityMetadata identifies missing or nonchronological lifecycle times.
var ErrInvalidEntityMetadata = errors.New("invalid entity metadata")

// EntityMetadata is an immutable value object describing an entity's lifetime.
type EntityMetadata struct {
	createdAt time.Time
	updatedAt time.Time
}

// NewEntityMetadata requires nonzero times with update at or after creation.
// Stored instants use UTC and omit monotonic clock readings.
func NewEntityMetadata(createdAt, updatedAt time.Time) (EntityMetadata, error) {
	if createdAt.IsZero() || updatedAt.IsZero() || updatedAt.Before(createdAt) {
		return EntityMetadata{}, fmt.Errorf("%w: nonzero times with updatedAt >= createdAt are required", ErrInvalidEntityMetadata)
	}
	return EntityMetadata{createdAt: createdAt.UTC().Round(0), updatedAt: updatedAt.UTC().Round(0)}, nil
}

// CreatedAt returns the original creation instant in UTC.
func (metadata EntityMetadata) CreatedAt() time.Time { return metadata.createdAt }

// UpdatedAt returns the most recent update instant in UTC.
func (metadata EntityMetadata) UpdatedAt() time.Time { return metadata.updatedAt }

// IsZero reports whether lifecycle metadata is uninitialized.
func (metadata EntityMetadata) IsZero() bool { return metadata.createdAt.IsZero() }

// Updated returns a copy with a later or equal update time.
func (metadata EntityMetadata) Updated(at time.Time) (EntityMetadata, error) {
	if metadata.IsZero() || at.IsZero() || at.Before(metadata.updatedAt) {
		return EntityMetadata{}, fmt.Errorf("%w: update time must not precede the previous update", ErrInvalidEntityMetadata)
	}
	return NewEntityMetadata(metadata.createdAt, at)
}
