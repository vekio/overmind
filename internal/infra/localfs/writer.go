// Package localfs persists rendered notes in the local filesystem.
package localfs

import (
	"context"
	"fmt"
	"path/filepath"

	"git.casta.me/alberto/overmind/internal/ports"
	"github.com/vekio/x/file"
	"uuid"
)

var _ ports.NoteWriter = (*Writer)(nil)

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
