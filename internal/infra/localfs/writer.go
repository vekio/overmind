// Package localfs persists rendered notes in the local filesystem.
package localfs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"git.casta.me/alberto/overmind/internal/ports"
	"github.com/vekio/x/file"
	"uuid"
)

var _ ports.NoteWriter = (*Writer)(nil)
var _ ports.NoteReader = (*Writer)(nil)
var _ ports.NoteDeleter = (*Writer)(nil)

// Writer stores notes in a directory.
type Writer struct {
	root string
}

// New creates a Writer rooted at root.
func New(root string) *Writer {
	return &Writer{root: root}
}

// Write stores content as <id>.adoc and returns its path.
func (writer *Writer) Write(ctx context.Context, id uuid.UUID, content []byte) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if id == uuid.Nil() {
		return "", fmt.Errorf("write note: UUID must not be nil")
	}
	if writer.root == "" {
		return "", fmt.Errorf("write note: storage root must not be empty")
	}

	path := filepath.Join(writer.root, id.String()+".adoc")
	if err := file.EnsureParentDir(path, 0o755); err != nil {
		return "", fmt.Errorf("create note directory: %w", err)
	}
	if err := file.WriteAtomic(path, content, 0o644); err != nil {
		return "", fmt.Errorf("write note %q: %w", path, err)
	}

	return path, nil
}

// Read returns the source of a note stored under its identifier.
func (writer *Writer) Read(ctx context.Context, id uuid.UUID) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id == uuid.Nil() || writer.root == "" {
		return nil, fmt.Errorf("read note: valid ID and storage root are required")
	}
	path := filepath.Join(writer.root, id.String()+".adoc")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read note %q: %w", path, err)
	}
	return content, nil
}

// Delete removes a note document. A missing file is already deleted.
func (writer *Writer) Delete(ctx context.Context, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id == uuid.Nil() || writer.root == "" {
		return fmt.Errorf("delete note: valid ID and storage root are required")
	}
	path := filepath.Join(writer.root, id.String()+".adoc")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete note %q: %w", path, err)
	}
	return nil
}
