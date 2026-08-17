// Package cli implements Overmind's command-line input adapter.
package cli

import (
	"context"

	"git.casta.me/alberto/overmind/internal/config"
	asciidoccli "git.casta.me/alberto/overmind/pkg/asciidoc/cli"
	urfavecli "github.com/urfave/cli/v3"
)

// NewCommand creates the complete overmind command tree.
func NewCommand(build func(config.Config) (Runtime, error)) *urfavecli.Command {
	state := &applicationState{}

	return &urfavecli.Command{
		Name:                  "overmind",
		Usage:                 "manage the overmind knowledge base",
		EnableShellCompletion: true,
		Flags:                 []urfavecli.Flag{configFlag(), debugFlag()},
		Before:                prepareApplication(state, build),
		After: func(_ context.Context, _ *urfavecli.Command) error {
			return state.close()
		},
		Commands: []*urfavecli.Command{
			asciidoccli.Command(),
			newEditCommand(state),
			newIndexCommand(state),
			newListCommand(state),
			newPageCommand(state),
			newServeCommand(state),
			newShowCommand(state),
		},
	}
}
