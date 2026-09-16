// Package bootstrap builds Overmind's dependency graph.
package bootstrap

import (
	"io"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/app/listdocuments"
	"git.casta.me/alberto/overmind/internal/app/rebuildindex"
	"git.casta.me/alberto/overmind/internal/app/updatedocument"
	"git.casta.me/alberto/overmind/internal/config"
	"git.casta.me/alberto/overmind/internal/ports"
)

// Runtime is the configured local application runtime.
type Runtime struct {
	app    *app.Application
	closer io.Closer
}

// Application returns the use cases exposed by the configured runtime.
func (runtime *Runtime) Application() *app.Application {
	return runtime.app
}

// Close releases resources owned by the runtime.
func (runtime *Runtime) Close() error {
	if runtime.closer == nil {
		return nil
	}
	return runtime.closer.Close()
}

type localDeps struct {
	Blobs    ports.BlobStore
	Renderer ports.Renderer
	IDs      ports.IDGenerator
	Clock    ports.Clock
	Index    ports.DocumentIndex
	Closer   io.Closer
}

// New builds the local application and all of its dependencies.
func New(cfg config.Config) (*Runtime, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	deps, err := newLocalDeps(cfg)
	if err != nil {
		return nil, err
	}

	return &Runtime{
		app:    newLocalApplication(deps),
		closer: deps.Closer,
	}, nil
}

func newLocalDeps(cfg config.Config) (localDeps, error) {
	blobs, err := newBlobStore(cfg.Vault)
	if err != nil {
		return localDeps{}, err
	}

	renderer, err := newRenderer()
	if err != nil {
		return localDeps{}, err
	}
	index, err := newDocumentIndex(cfg.Index)
	if err != nil {
		return localDeps{}, err
	}

	return localDeps{
		Blobs:    blobs,
		Renderer: renderer,
		IDs:      newIDGenerator(),
		Clock:    newClock(),
		Index:    index,
		Closer:   index,
	}, nil
}

func newLocalApplication(deps localDeps) *app.Application {
	return &app.Application{
		Commands: app.Commands{
			CreatePage: createpage.NewCreatePageHandler(
				deps.Blobs, deps.Renderer, deps.IDs, deps.Clock, deps.Index,
			),
			RebuildIndex: rebuildindex.NewRebuildIndexHandler(deps.Blobs, deps.Index),
			UpdateDocument: updatedocument.NewUpdateDocumentHandler(
				deps.Blobs, deps.Blobs, deps.Clock, deps.Index,
			),
		},
		Queries: app.Queries{
			GetDocument:   getdocument.NewGetDocumentHandler(deps.Blobs),
			ListDocuments: listdocuments.NewListDocumentsHandler(deps.Index),
		},
	}
}
