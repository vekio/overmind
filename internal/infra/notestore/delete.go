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
)

// Delete resolves identity throughout the managed notes directory, including
// subdirectories visited by Scan. Duplicate identities are rejected before
// mutation so removing an index row cannot hide another source document.
func (store *Store) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := store.path(ctx, id); err != nil {
		return err
	}
	var matches []string
	err := filepath.WalkDir(store.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if path == store.root && errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if filepath.Ext(entry.Name()) != ".adoc" {
			return nil
		}
		candidate, err := uuid.Parse(strings.TrimSuffix(entry.Name(), ".adoc"))
		if err != nil || candidate != id {
			return nil
		}
		if entry.IsDir() {
			return fmt.Errorf("%s: note path is a directory", path)
		}
		matches = append(matches, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("locate note for deletion: %w", err)
	}
	if len(matches) > 1 {
		return fmt.Errorf("duplicate note ID %s in %s", id, strings.Join(matches, ", "))
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(matches) == 0 {
		return nil
	}
	if err := os.Remove(matches[0]); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("delete note: %w", err)
	}
	return nil
}
