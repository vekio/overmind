package notestore

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"uuid"

	"github.com/vekio/overmind/internal/ports"
)

var _ ports.NoteScanner = (*Store)(nil)

// Scan reads managed .adoc files recursively, in filename order. Filenames
// carry their UUID, as in Put and Get; other file extensions are ignored.
func (store *Store) Scan(ctx context.Context, visit func(ports.Note) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if store.root == "" || visit == nil {
		return fmt.Errorf("note storage root and visitor are required")
	}
	if info, err := os.Stat(store.root); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("%s: note storage root must be a directory", store.root)
	}
	return filepath.WalkDir(store.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".adoc" {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("%s: note must be a regular file", path)
		}
		id, err := uuid.Parse(strings.TrimSuffix(entry.Name(), ".adoc"))
		if err != nil || id == uuid.Nil() {
			return fmt.Errorf("%s: note filename must be a nonnil UUID", path)
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		return visit(ports.Note{ID: id, Path: path, Content: source})
	})
}
