package cli

import (
	urfavecli "github.com/urfave/cli/v3"
)

func New(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:                  "overmind",
		Usage:                 "manage the overmind knowledge base",
		EnableShellCompletion: true,
		Commands: []*urfavecli.Command{
			newBookmarkCommand(newClient),
			newCaptureCommand(newClient),
			newJournalCommand(newClient),
			newPageCommand(newClient),
		},
	}
}
