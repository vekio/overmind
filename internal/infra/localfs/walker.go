package localfs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vekio/overmind/internal/ports"
)

var _ ports.NoteWalker = (*Walker)(nil)

// Walker visits regular AsciiDoc notes under a directory.
type Walker struct {
	root string
}

// NewWalker creates a walker rooted at the notes directory.
func NewWalker(root string) *Walker {
	return &Walker{root: root}
}

// Walk reads each regular .adoc file and passes it to visit.
func (walker *Walker) Walk(ctx context.Context, visit func(path string, content []byte) error) error {
	info, err := os.Stat(walker.root)
	if err != nil {
		return fmt.Errorf("open notes directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("notes path %q is not a directory", walker.root)
	}

	err = filepath.WalkDir(walker.root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() || !entry.Type().IsRegular() || !strings.EqualFold(filepath.Ext(path), ".adoc") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read note %q: %w", path, err)
		}
		return visit(path, content)
	})
	if err != nil {
		return fmt.Errorf("walk notes: %w", err)
	}
	return nil
}
