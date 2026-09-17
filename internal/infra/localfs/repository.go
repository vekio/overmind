// Package localfs stores Overmind documents in a local directory.
package localfs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/vekio/x/file"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

var _ ports.BlobStore = (*Store)(nil)

// Store persists each document under its stable ID.
type Store struct {
	rootPath string
}

// New creates a local document store rooted at rootPath.
func New(rootPath string) *Store {
	return &Store{rootPath: rootPath}
}

func (store *Store) Create(ctx context.Context, blob ports.Blob) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	documentPath, err := store.documentPath(blob.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(store.rootPath, 0o755); err != nil {
		return err
	}
	if err := file.WriteExclusive(documentPath, blob.Content, 0o644); errors.Is(err, fs.ErrExist) {
		return ports.ErrBlobAlreadyExists
	} else {
		return err
	}
}

func (store *Store) Put(ctx context.Context, blob ports.Blob) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	documentPath, err := store.documentPath(blob.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(store.rootPath, 0o755); err != nil {
		return err
	}
	return file.WriteAtomic(documentPath, blob.Content, 0o644)
}

// Update replaces an existing document when its current revision matches the
// expected revision.
func (store *Store) Update(ctx context.Context, blob ports.Blob, expectedRevision string) (string, error) {
	current, err := store.Get(ctx, blob.ID)
	if err != nil {
		return "", err
	}
	if current.Revision != expectedRevision {
		return "", ports.ErrBlobChanged
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	documentPath, err := store.documentPath(blob.ID)
	if err != nil {
		return "", err
	}

	// Recheck immediately before replacement so a concurrent edit is not
	// knowingly overwritten.
	currentContent, err := os.ReadFile(documentPath)
	if errors.Is(err, fs.ErrNotExist) {
		return "", ports.ErrBlobNotFound
	}
	if err != nil {
		return "", err
	}
	if blobRevision(currentContent) != expectedRevision {
		return "", ports.ErrBlobChanged
	}
	if err := file.WriteAtomic(documentPath, blob.Content, 0o644); err != nil {
		return "", err
	}
	return blobRevision(blob.Content), nil
}

func (store *Store) Get(ctx context.Context, id domain.DocumentID) (ports.Blob, error) {
	if err := ctx.Err(); err != nil {
		return ports.Blob{}, err
	}
	documentPath, err := store.documentPath(id)
	if err != nil {
		return ports.Blob{}, err
	}
	content, err := os.ReadFile(documentPath)
	if errors.Is(err, fs.ErrNotExist) {
		return ports.Blob{}, ports.ErrBlobNotFound
	}
	if err != nil {
		return ports.Blob{}, err
	}
	return ports.Blob{ID: id, Content: content, Revision: blobRevision(content)}, nil
}

func (store *Store) Delete(ctx context.Context, id domain.DocumentID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	documentPath, err := store.documentPath(id)
	if err != nil {
		return err
	}
	if err := os.Remove(documentPath); errors.Is(err, fs.ErrNotExist) {
		return ports.ErrBlobNotFound
	} else {
		return err
	}
}

func (store *Store) List(ctx context.Context) ([]domain.DocumentID, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(store.rootPath)
	if errors.Is(err, fs.ErrNotExist) {
		return []domain.DocumentID{}, nil
	}
	if err != nil {
		return nil, err
	}

	ids := make([]domain.DocumentID, 0, len(entries))
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !entry.Type().IsRegular() || filepath.Ext(entry.Name()) != ".adoc" {
			continue
		}
		id, err := domain.NewDocumentID(strings.TrimSuffix(entry.Name(), ".adoc"))
		if err == nil {
			ids = append(ids, id)
		}
	}
	slices.SortFunc(ids, func(left, right domain.DocumentID) int {
		return strings.Compare(left.String(), right.String())
	})
	return ids, nil
}

func (store *Store) documentPath(id domain.DocumentID) (string, error) {
	validated, err := domain.NewDocumentID(id.String())
	if err != nil {
		return "", err
	}
	return filepath.Join(store.rootPath, validated.String()+".adoc"), nil
}

func blobRevision(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}
