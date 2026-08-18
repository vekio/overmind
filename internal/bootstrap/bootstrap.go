// Package bootstrap builds Overmind's dependency graph.
package bootstrap

import (
	"io"
	"log/slog"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/app/listdocuments"
	appmiddleware "git.casta.me/alberto/overmind/internal/app/middleware"
	"git.casta.me/alberto/overmind/internal/app/rebuildindex"
	"git.casta.me/alberto/overmind/internal/app/updatedocument"
	"git.casta.me/alberto/overmind/internal/config"
	"git.casta.me/alberto/overmind/internal/ports"
)

// Container is the configured application runtime.
type Container struct {
	log    *slog.Logger
	app    *app.Application
	closer io.Closer
}

type runtimeConfig struct {
	Vault   config.Vault
	Logging config.Logging
	Index   config.Index
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

// NewCLIContainer builds the local CLI dependency graph.
func NewCLIContainer(cfg config.CLIConfig) (*Container, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return newContainer(runtimeConfig{Vault: cfg.Vault, Logging: cfg.Logging, Index: cfg.Index})
}

// NewServerContainer builds the HTTP server dependency graph.
func NewServerContainer(cfg config.ServerConfig) (*Container, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return newContainer(runtimeConfig{Vault: cfg.Vault, Logging: cfg.Logging, Index: cfg.Index})
}

func newContainer(cfg runtimeConfig) (*Container, error) {
	log, err := newLogger(cfg.Logging)
	if err != nil {
		return nil, err
	}

	deps, err := newLocalDeps(cfg)
	if err != nil {
		return nil, err
	}

	return &Container{
		log:    log,
		app:    newLocalApplication(deps, log),
		closer: deps.Closer,
	}, nil
}

func newLocalDeps(cfg runtimeConfig) (localDeps, error) {
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

func newLocalApplication(deps localDeps, logger *slog.Logger) *app.Application {
	applicationLogger := logger.With("component", "application")
	createPageHandler := createpage.NewCreatePageHandler(
		deps.Blobs,
		deps.Renderer,
		deps.IDs,
		deps.Clock,
		deps.Index,
	)
	getDocumentHandler := getdocument.NewGetDocumentHandler(deps.Blobs)
	updateDocumentHandler := updatedocument.NewUpdateDocumentHandler(deps.Blobs, deps.Blobs, deps.Clock, deps.Index)
	listDocumentsHandler := listdocuments.NewListDocumentsHandler(deps.Index)
	rebuildIndexHandler := rebuildindex.NewRebuildIndexHandler(deps.Blobs, deps.Index)

	return &app.Application{
		Commands: app.Commands{
			CreatePage: appmiddleware.Logging(
				"create_page",
				createPageHandler,
				applicationLogger,
			),
			RebuildIndex: appmiddleware.Logging(
				"rebuild_index",
				rebuildIndexHandler,
				applicationLogger,
			),
			UpdateDocument: appmiddleware.Logging(
				"update_document",
				updateDocumentHandler,
				applicationLogger,
			),
		},
		Queries: app.Queries{
			GetDocument: appmiddleware.Logging(
				"get_document",
				getDocumentHandler,
				applicationLogger,
			),
			ListDocuments: appmiddleware.Logging(
				"list_documents",
				listDocumentsHandler,
				applicationLogger,
			),
		},
	}
}
