package cli

import (
	"git.casta.me/alberto/overmind/internal/renderer"
	urfavecli "github.com/urfave/cli/v3"
)

func New() (*urfavecli.Command, error) {
	documentRenderer, err := renderer.New()
	if err != nil {
		return nil, err
	}

	return &urfavecli.Command{
		Name:                  "overmind",
		Usage:                 "manage the overmind knowledge base",
		EnableShellCompletion: true,
		Commands: []*urfavecli.Command{
			newPageCommand(documentRenderer),
		},
	}, nil
}
