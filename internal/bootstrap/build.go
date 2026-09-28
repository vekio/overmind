package bootstrap

import (
	"context"
	"io"
	"path/filepath"

	"git.casta.me/alberto/overmind/internal/app"
	appconfig "git.casta.me/alberto/overmind/internal/config"
	"git.casta.me/alberto/overmind/internal/infra/asciidocnote"
	"git.casta.me/alberto/overmind/internal/infra/idgenerator"
	"git.casta.me/alberto/overmind/internal/infra/localfs"
	"git.casta.me/alberto/overmind/internal/infra/renderer"
	"git.casta.me/alberto/overmind/internal/infra/sqliteindex"
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
