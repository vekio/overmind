package ports

import (
	"context"
	"errors"
)

var (
	ErrBlobAlreadyExists = errors.New("blob already exists")
	ErrBlobNotFound      = errors.New("blob not found")
	ErrBlobChanged       = errors.New("blob changed since it was read")
	ErrInvalidBlobPath   = errors.New("invalid blob path")
)

// Blob is binary content identified by a store-relative path.
type Blob struct {
	Path    string
	Content []byte
	// Revision is an opaque value that changes with the blob content.
	Revision string
}

// BlobFilter restricts blobs returned by BlobReader.List.
type BlobFilter struct {
	Prefix string
	Suffix string
}

// BlobWriter is the persistence port used by commands.
type BlobWriter interface {
	Create(ctx context.Context, blob Blob) error
	Put(ctx context.Context, blob Blob) error
	Delete(ctx context.Context, blobPath string) error
}

// BlobUpdater conditionally replaces an existing blob. The returned revision
// identifies the newly stored content.
type BlobUpdater interface {
	Update(ctx context.Context, blob Blob, expectedRevision string) (string, error)
}

// BlobReader is the persistence port used by queries.
type BlobReader interface {
	Get(ctx context.Context, blobPath string) (Blob, error)
	List(ctx context.Context, filter BlobFilter) ([]string, error)
}

// BlobStore combines the blob read and write ports for adapters that implement
// both. Application handlers should depend on BlobReader or BlobWriter.
type BlobStore interface {
	BlobReader
	BlobWriter
	BlobUpdater
}
