package ports

import "uuid"

// IDGenerator generates entity identifiers.
type IDGenerator interface{ Generate() uuid.UUID }
