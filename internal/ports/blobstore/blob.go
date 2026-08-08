// Package blobstore defines the port for persistent blob storage.
package blobstore

// Blob is binary content identified by a repository-relative path.
type Blob struct {
	Path    string
	Content []byte
}
