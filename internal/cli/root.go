package cli

import (
	"git.casta.me/alberto/overmind/internal/app"
	urfavecli "github.com/urfave/cli/v3"
)

func New(application *app.App) *urfavecli.Command {
	return &urfavecli.Command{
		Name:                  "overmind",
		Usage:                 "manage the overmind knowledge base",
		EnableShellCompletion: true,
		Commands: []*urfavecli.Command{
			newBookmarkCommand(application),
			newCaptureCommand(application),
			newJournalCommand(application),
			newPageCommand(application),
		},
	}
}
