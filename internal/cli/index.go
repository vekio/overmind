package cli

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/app/rebuildindex"
	"git.casta.me/alberto/overmind/internal/config"
	urfavecli "github.com/urfave/cli/v3"
)

func newIndexCommand(state *applicationState) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "index",
		Usage: "manage the local document index",
		Commands: []*urfavecli.Command{
			{
				Name:  "rebuild",
				Usage: "rebuild the index from the AsciiDoc vault",
				Action: func(ctx context.Context, command *urfavecli.Command) error {
					if state.config.CLI.Mode != config.CLIModeLocal {
						return fmt.Errorf("index rebuild requires cli.mode %q", config.CLIModeLocal)
					}
					application, err := state.get()
					if err != nil {
						return err
					}
					if application.Commands.RebuildIndex == nil {
						return errMissingRuntime
					}
					result, err := application.Commands.RebuildIndex.Handle(ctx, rebuildindex.RebuildIndexCommand{})
					if err != nil {
						return err
					}
					_, err = fmt.Fprintf(command.Writer, "Indexed %d documents\n", result.Documents)
					return err
				},
			},
		},
	}
}
