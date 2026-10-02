package bootstrap

import (
	"context"
	"io"
	"path/filepath"

	"github.com/vekio/overmind/internal/app"
	appconfig "github.com/vekio/overmind/internal/config"
	"github.com/vekio/overmind/internal/infra/asciidocnote"
	"github.com/vekio/overmind/internal/infra/idgenerator"
	"github.com/vekio/overmind/internal/infra/localfs"
	"github.com/vekio/overmind/internal/infra/renderer"
	"github.com/vekio/overmind/internal/infra/sqliteindex"
)

func buildApplication(
	ctx context.Context,
	settings appconfig.Settings,
) (*app.Application, []io.Closer, error) {
	documentRenderer, err := renderer.New()
	if err != nil {
		return nil, nil, err
	}

	index, err := sqliteindex.New(ctx, filepath.Join(settings.VaultPath, "index.db"))
	if err != nil {
		return nil, nil, err
	}
	notesPath := filepath.Join(settings.VaultPath, "notes")
	noteFiles := localfs.New(notesPath)

	application := app.New(app.Dependencies{
		IDGenerator: idgenerator.New(),
		Renderer:    documentRenderer,
		Writer:      noteFiles,
		Reader:      noteFiles,
		Deleter:     noteFiles,
		Index:       index,
		Lister:      index,
		Walker:      localfs.NewWalker(notesPath),
		Parser:      asciidocnote.Parser{},
	})
	// Add future closable resources in creation order.
	return application, []io.Closer{index}, nil
}
