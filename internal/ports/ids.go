package ports

import "uuid"

// IDGenerator generates entity identifiers.
type IDGenerator interface {
	// Generate returns a fresh UUID for a new entity.
	Generate() uuid.UUID
}
