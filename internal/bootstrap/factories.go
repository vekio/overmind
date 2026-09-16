package bootstrap

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/config"
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

func newBlobStore(cfg config.Vault) (ports.BlobStore, error) {
	switch cfg.Driver {
	case "local":
		return localfs.New(cfg.RootPath), nil
	default:
		return nil, fmt.Errorf("unsupported vault driver %q", cfg.Driver)
	}
}

func newDocumentIndex(cfg config.Index) (*sqliteindex.Store, error) {
	switch cfg.Driver {
	case "sqlite":
		return sqliteindex.New(context.Background(), cfg.Path)
	default:
		return nil, fmt.Errorf("unsupported index driver %q", cfg.Driver)
	}
}

func newRenderer() (ports.Renderer, error) {
	return gotemplate.New()
}
