package blobstore

import "context"

// Store persists and retrieves blobs.
type Store interface {
	// Create stores a new blob and fails if the path already exists.
	Create(ctx context.Context, blob Blob) error

	// Put stores a blob, replacing any existing blob at the same path.
	Put(ctx context.Context, blob Blob) error

	// Get returns the blob stored at path.
	Get(ctx context.Context, blobPath string) (Blob, error)

	// Delete removes the blob stored at path.
	Delete(ctx context.Context, blobPath string) error

	// List returns blob paths that match filter.
	List(ctx context.Context, filter Filter) ([]string, error)
}

// Filter restricts the paths returned by Store.List.
type Filter struct {
	Prefix string
	Suffix string
}
