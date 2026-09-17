package bootstrap

import (
	"git.casta.me/alberto/overmind/internal/config"
	"git.casta.me/alberto/overmind/internal/infra/sqliteindex"
	"git.casta.me/alberto/overmind/internal/ports"
)

type deps struct {
	blobs    ports.BlobStore
	renderer ports.Renderer
	ids      ports.IDGenerator
	clock    ports.Clock
	index    *sqliteindex.Store
}

func newDeps(cfg config.Config) (deps, error) {
	if err := cfg.Validate(); err != nil {
		return deps{}, err
	}
	renderer, err := newRenderer()
	if err != nil {
		return deps{}, err
	}
	index, err := newDocumentIndex(cfg.DataDir)
	if err != nil {
		return deps{}, err
	}
	return deps{
		blobs:    newBlobStore(cfg.DataDir),
		renderer: renderer,
		ids:      newIDGenerator(),
		clock:    newClock(),
		index:    index,
	}, nil
}
