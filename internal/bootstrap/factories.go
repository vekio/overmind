package bootstrap

import (
	"context"
	"path/filepath"

	"git.casta.me/alberto/overmind/internal/infra/gotemplate"
	"git.casta.me/alberto/overmind/internal/infra/localfs"
	"git.casta.me/alberto/overmind/internal/infra/sqliteindex"
	"git.casta.me/alberto/overmind/internal/infra/systemclock"
	"git.casta.me/alberto/overmind/internal/infra/uuidgenerator"
	"git.casta.me/alberto/overmind/internal/ports"
)

func newClock() ports.Clock {
	return systemclock.New()
}

func newIDGenerator() ports.IDGenerator {
	return uuidgenerator.New()
}

func newBlobStore(dataDir string) ports.BlobStore {
	return localfs.New(filepath.Join(dataDir, "documents"))
}

func newDocumentIndex(dataDir string) (*sqliteindex.Store, error) {
	return sqliteindex.New(context.Background(), filepath.Join(dataDir, "index.db"))
}

func newRenderer() (ports.Renderer, error) {
	return gotemplate.New()
}
