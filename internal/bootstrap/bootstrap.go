// Package bootstrap builds Overmind's dependency graph.
package bootstrap

import (
	"fmt"
	"io"
	"log/slog"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/app/rebuildindex"
	"git.casta.me/alberto/overmind/internal/config"
	"git.casta.me/alberto/overmind/internal/httpclient"
	httpclientdocuments "git.casta.me/alberto/overmind/internal/httpclient/documents"
	httpclientpages "git.casta.me/alberto/overmind/internal/httpclient/pages"
	"git.casta.me/alberto/overmind/internal/ports"
)

// Container is the configured application runtime.
type Container struct {
	config config.Config
	log    *slog.Logger
	app    *app.Application
	closer io.Closer
}

// Application returns the use cases exposed by the configured runtime.
func (container *Container) Application() *app.Application {
	return container.app
}

// Logger returns the process logger shared by the input adapters.
func (container *Container) Logger() *slog.Logger {
	return container.log
}

// Close releases infrastructure resources owned by the container.
func (container *Container) Close() error {
	if container.closer == nil {
		return nil
	}
	return container.closer.Close()
}

type localDeps struct {
	Blobs    ports.BlobStore
	Renderer ports.Renderer
	IDs      ports.IDGenerator
	Clock    ports.Clock
	Index    ports.DocumentIndex
	Closer   io.Closer
}

// NewContainer builds the dependency graph for a configuration.
func NewContainer(cfg config.Config) (*Container, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	log, err := newLogger(cfg.Logging)
	if err != nil {
		return nil, err
	}

	application, closer, err := newConfiguredApplication(cfg)
	if err != nil {
		return nil, err
	}

	return &Container{
		config: cfg,
		log:    log,
		app:    application,
		closer: closer,
	}, nil
}

func newConfiguredApplication(cfg config.Config) (*app.Application, io.Closer, error) {
	switch cfg.CLI.Mode {
	case config.CLIModeLocal:
		deps, err := newLocalDeps(cfg)
		if err != nil {
			return nil, nil, err
		}
		return newLocalApplication(deps), deps.Closer, nil
	case config.CLIModeRemote:
		client, err := newOvermindClient(cfg.CLI)
		if err != nil {
			return nil, nil, err
		}
		return newRemoteApplication(client), nil, nil
	default:
		return nil, nil, fmt.Errorf("unsupported CLI mode %q", cfg.CLI.Mode)
	}
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
				deps.Blobs,
				deps.Renderer,
				deps.IDs,
				deps.Clock,
				deps.Index,
			),
			RebuildIndex: rebuildindex.NewRebuildIndexHandler(deps.Blobs, deps.Index),
		},
		Queries: app.Queries{
			GetDocument: getdocument.NewGetDocumentHandler(deps.Index),
		},
	}
}

func newRemoteApplication(client *httpclient.Client) *app.Application {
	return &app.Application{
		Commands: app.Commands{
			CreatePage: httpclientpages.NewCreateHandler(client),
		},
		Queries: app.Queries{
			GetDocument: httpclientdocuments.NewGetHandler(client),
		},
	}
}
