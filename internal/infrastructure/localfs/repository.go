package localfs

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"git.casta.me/alberto/overmind/internal/ports"
)

var _ ports.BlobStore = (*Store)(nil)

// Store implements blob storage in a local directory.
type Store struct {
	rootPath string
}

// New creates a local filesystem blob store rooted at rootPath.
func New(rootPath string) *Store {
	return &Store{rootPath: rootPath}
}

func (store *Store) Create(ctx context.Context, blob ports.Blob) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	blobPath, err := normalizeBlobPath(blob.Path)
	if err != nil {
		return err
	}

	root, err := store.openRepositoryRoot(true)
	if err != nil {
		return err
	}
	defer root.Close()

	localPath := filepath.FromSlash(blobPath)
	if err := root.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return err
	}

	targetFile, err := root.OpenFile(localPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return ports.ErrBlobAlreadyExists
	}
	if err != nil {
		return err
	}
	defer targetFile.Close()

	written, err := targetFile.Write(blob.Content)
	if err != nil {
		return err
	}
	if written != len(blob.Content) {
		return io.ErrShortWrite
	}

	return nil
}

func (store *Store) Put(ctx context.Context, blob ports.Blob) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	blobPath, err := normalizeBlobPath(blob.Path)
	if err != nil {
		return err
	}

	root, err := store.openRepositoryRoot(true)
	if err != nil {
		return err
	}
	defer root.Close()

	localPath := filepath.FromSlash(blobPath)
	if err := root.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return err
	}

	return root.WriteFile(localPath, blob.Content, 0o644)
}

// Update replaces an existing blob when its current revision matches the
// expected revision. A same-directory rename keeps the visible replacement
// atomic for readers.
func (store *Store) Update(ctx context.Context, blob ports.Blob, expectedRevision string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	blobPath, err := normalizeBlobPath(blob.Path)
	if err != nil {
		return "", err
	}
	root, err := store.openRepositoryRoot(false)
	if errors.Is(err, fs.ErrNotExist) {
		return "", ports.ErrBlobNotFound
	}
	if err != nil {
		return "", err
	}
	defer root.Close()

	localPath := filepath.FromSlash(blobPath)
	currentContent, err := root.ReadFile(localPath)
	if errors.Is(err, fs.ErrNotExist) {
		return "", ports.ErrBlobNotFound
	}
	if err != nil {
		return "", err
	}
	if blobRevision(currentContent) != expectedRevision {
		return "", ports.ErrBlobChanged
	}

	temporaryPath, err := writeTemporaryBlob(root, localPath, blob.Content)
	if err != nil {
		return "", err
	}
	defer func() { _ = root.Remove(temporaryPath) }()

	if err := ctx.Err(); err != nil {
		return "", err
	}
	// Recheck immediately before the rename so an edit made while the temporary
	// file was being written is not knowingly overwritten.
	currentContent, err = root.ReadFile(localPath)
	if errors.Is(err, fs.ErrNotExist) {
		return "", ports.ErrBlobNotFound
	}
	if err != nil {
		return "", err
	}
	if blobRevision(currentContent) != expectedRevision {
		return "", ports.ErrBlobChanged
	}
	if err := root.Rename(temporaryPath, localPath); err != nil {
		return "", err
	}
	return blobRevision(blob.Content), nil
}

func (store *Store) Get(ctx context.Context, blobPath string) (ports.Blob, error) {
	if err := ctx.Err(); err != nil {
		return ports.Blob{}, err
	}

	cleanPath, err := normalizeBlobPath(blobPath)
	if err != nil {
		return ports.Blob{}, err
	}

	root, err := store.openRepositoryRoot(false)
	if errors.Is(err, fs.ErrNotExist) {
		return ports.Blob{}, ports.ErrBlobNotFound
	}
	if err != nil {
		return ports.Blob{}, err
	}
	defer root.Close()

	contentBytes, err := root.ReadFile(filepath.FromSlash(cleanPath))
	if errors.Is(err, fs.ErrNotExist) {
		return ports.Blob{}, ports.ErrBlobNotFound
	}
	if err != nil {
		return ports.Blob{}, err
	}

	return ports.Blob{
		Path:     cleanPath,
		Content:  contentBytes,
		Revision: blobRevision(contentBytes),
	}, nil
}

func (store *Store) Delete(ctx context.Context, blobPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	cleanPath, err := normalizeBlobPath(blobPath)
	if err != nil {
		return err
	}

	root, err := store.openRepositoryRoot(false)
	if errors.Is(err, fs.ErrNotExist) {
		return ports.ErrBlobNotFound
	}
	if err != nil {
		return err
	}
	defer root.Close()

	if err := root.Remove(filepath.FromSlash(cleanPath)); errors.Is(err, fs.ErrNotExist) {
		return ports.ErrBlobNotFound
	} else if err != nil {
		return err
	}

	return nil
}

func (store *Store) List(ctx context.Context, filter ports.BlobFilter) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	root, err := store.openRepositoryRoot(false)
	if errors.Is(err, fs.ErrNotExist) {
		return []string{}, nil
	} else if err != nil {
		return nil, err
	}
	defer root.Close()

	paths := []string{}
	err = fs.WalkDir(root.FS(), ".", func(blobPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}

			return nil
		}

		// Directory walking does not follow symlinks. Omitting them keeps List
		// consistent with the repository's root-confined read behavior.
		if entry.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		blobPath, err = normalizeBlobPath(blobPath)
		if err != nil {
			return nil
		}

		if filter.Prefix != "" && !strings.HasPrefix(blobPath, filter.Prefix) {
			return nil
		}
		if filter.Suffix != "" && !strings.HasSuffix(blobPath, filter.Suffix) {
			return nil
		}

		paths = append(paths, blobPath)
		return nil
	})
	if err != nil {
		return nil, err
	}

	slices.Sort(paths)

	return paths, nil
}

func (store *Store) openRepositoryRoot(create bool) (*os.Root, error) {
	if create {
		if err := os.MkdirAll(store.rootPath, 0o755); err != nil {
			return nil, err
		}
	}

	return os.OpenRoot(store.rootPath)
}

func normalizeBlobPath(blobPath string) (string, error) {
	blobPath = strings.TrimSpace(blobPath)
	if blobPath == "" || strings.Contains(blobPath, "\\") {
		return "", ports.ErrInvalidBlobPath
	}

	clean := path.Clean(blobPath)
	if clean == "." || path.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", ports.ErrInvalidBlobPath
	}
	return clean, nil
}

func blobRevision(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func writeTemporaryBlob(root *os.Root, targetPath string, content []byte) (string, error) {
	var randomBytes [8]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", err
	}
	temporaryPath := targetPath + ".overmind-" + hex.EncodeToString(randomBytes[:]) + ".tmp"
	file, err := root.OpenFile(temporaryPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}

	written, writeErr := file.Write(content)
	if writeErr == nil && written != len(content) {
		writeErr = io.ErrShortWrite
	}
	if writeErr == nil {
		writeErr = file.Sync()
	}
	if closeErr := file.Close(); writeErr == nil {
		writeErr = closeErr
	}
	if writeErr != nil {
		_ = root.Remove(temporaryPath)
		return "", writeErr
	}
	return temporaryPath, nil
}
