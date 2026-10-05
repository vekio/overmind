// Package notestore persists note documents atomically in the filesystem.
package notestore

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"uuid"

	"github.com/vekio/overmind/internal/ports"
	"github.com/vekio/x/file"
)

var _ ports.NoteStore = (*Store)(nil)

// Store maps UUIDs to managed documents under one filesystem root.
type Store struct{ root string }

// New binds a storage root without creating directories or reading documents.
func New(root string) *Store { return &Store{root: root} }
func (store *Store) path(ctx context.Context, id uuid.UUID) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if id == uuid.Nil() || store.root == "" {
		return "", fmt.Errorf("note ID and storage root are required")
	}
	return filepath.Join(store.root, id.String()+".adoc"), nil
}

// Put atomically writes exact source bytes to the canonical UUID filename.
// The supplied Path is ignored; the returned path identifies the persisted file.
func (store *Store) Put(ctx context.Context, note ports.Note) (string, error) {
	path, err := store.path(ctx, note.ID)
	if err != nil {
		return "", err
	}
	if err := file.EnsureParentDir(path, 0755); err != nil {
		return "", fmt.Errorf("create note directory: %w", err)
	}
	if err := file.WriteAtomic(path, note.Content, 0644); err != nil {
		return "", fmt.Errorf("write note: %w", err)
	}
	return path, nil
}

// Get reads the canonical UUID filename and preserves its bytes without decoding.
// Missing files return ports.ErrNoteNotFound.
func (store *Store) Get(ctx context.Context, id uuid.UUID) (ports.Note, error) {
	path, err := store.path(ctx, id)
	if err != nil {
		return ports.Note{}, err
	}
	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return ports.Note{}, fmt.Errorf("%w: %s", ports.ErrNoteNotFound, id)
	}
	if err != nil {
		return ports.Note{}, fmt.Errorf("read note: %w", err)
	}
	return ports.Note{ID: id, Content: content, Path: path}, nil
}
