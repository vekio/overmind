package cli

import (
	"git.casta.me/alberto/overmind/internal/renderer"
	"git.casta.me/alberto/overmind/internal/storage"
	urfavecli "github.com/urfave/cli/v3"
)

const defaultNotesRoot = ".overmind/notes"

func New() (*urfavecli.Command, error) {
	documentRenderer, err := renderer.New()
	if err != nil {
		return nil, err
	}
	noteWriter := storage.NewFileWriter(defaultNotesRoot)

	return &urfavecli.Command{
		Name:                  "overmind",
		Usage:                 "manage the overmind knowledge base",
		EnableShellCompletion: true,
		Commands: []*urfavecli.Command{
			newBookmarkCommand(documentRenderer, noteWriter),
			newCaptureCommand(documentRenderer, noteWriter),
			newJournalCommand(documentRenderer, noteWriter),
			newPageCommand(documentRenderer, noteWriter),
		},
	}, nil
}
