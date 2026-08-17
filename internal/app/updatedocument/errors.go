package updatedocument

import "errors"

var (
	// ErrRevisionRequired means the caller did not identify the version edited.
	ErrRevisionRequired = errors.New("document revision is required")
	// ErrUnmanagedDocument means the source has no valid Overmind metadata.
	ErrUnmanagedDocument = errors.New("document is not managed by Overmind")
	// ErrImmutableMetadata means editing changed stable document metadata.
	ErrImmutableMetadata = errors.New("document id, type, and creation time cannot be changed")
)
