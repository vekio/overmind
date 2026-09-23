package cli

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/infra/sqliteindex"
	"git.casta.me/alberto/overmind/internal/renderer"
	"git.casta.me/alberto/overmind/internal/storage"
	urfavecli "github.com/urfave/cli/v3"
)

const defaultNotesRoot = ".overmind/notes"
const defaultIndexPath = ".overmind/index.db"

func New() (*urfavecli.Command, error) {
	documentRenderer, err := renderer.New()
	if err != nil {
		return nil, err
	}
	noteWriter := storage.NewFileWriter(defaultNotesRoot)
	noteIndex, err := sqliteindex.New(context.Background(), defaultIndexPath)
	if err != nil {
		return nil, err
	}

	return &urfavecli.Command{
		Name:                  "overmind",
		Usage:                 "manage the overmind knowledge base",
		EnableShellCompletion: true,
		After: func(context.Context, *urfavecli.Command) error {
			if err := noteIndex.Close(); err != nil {
				return fmt.Errorf("close SQLite index: %w", err)
			}
			return nil
		},
		Commands: []*urfavecli.Command{
			newBookmarkCommand(documentRenderer, noteWriter, noteIndex),
			newCaptureCommand(documentRenderer, noteWriter, noteIndex),
			newJournalCommand(documentRenderer, noteWriter, noteIndex),
			newPageCommand(documentRenderer, noteWriter, noteIndex),
		},
	}, nil
}
