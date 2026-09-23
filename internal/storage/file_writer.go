// Package storage persists rendered Overmind notes.
package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"uuid"
)

// FileWriter stores notes in a directory.
type FileWriter struct {
	root string
}

// NewFileWriter creates a FileWriter rooted at root.
func NewFileWriter(root string) *FileWriter {
	return &FileWriter{root: root}
}

// Write stores content as <id>.adoc and returns its path.
func (writer *FileWriter) Write(ctx context.Context, id uuid.UUID, content []byte) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if id == uuid.Nil() {
		return "", fmt.Errorf("write note: UUID must not be nil")
	}
	if writer.root == "" {
		return "", fmt.Errorf("write note: storage root must not be empty")
	}
	if err := os.MkdirAll(writer.root, 0o755); err != nil {
		return "", fmt.Errorf("create storage root %q: %w", writer.root, err)
	}

	temporary, err := os.CreateTemp(writer.root, ".overmind-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temporary note: %w", err)
	}
	temporaryPath := temporary.Name()
	closed := false
	defer func() {
		if !closed {
			_ = temporary.Close()
		}
		_ = os.Remove(temporaryPath)
	}()

	if err := temporary.Chmod(0o644); err != nil {
		return "", fmt.Errorf("set temporary note permissions: %w", err)
	}
	if _, err := temporary.Write(content); err != nil {
		return "", fmt.Errorf("write temporary note: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return "", fmt.Errorf("sync temporary note: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("close temporary note: %w", err)
	}
	closed = true

	if err := ctx.Err(); err != nil {
		return "", err
	}
	path := filepath.Join(writer.root, id.String()+".adoc")
	if err := os.Link(temporaryPath, path); err != nil {
		return "", fmt.Errorf("publish note %q: %w", path, err)
	}

	return path, nil
}
