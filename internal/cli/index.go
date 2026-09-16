package cli

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/app/rebuildindex"
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
					application, err := state.Application()
					if err != nil {
						return err
					}
					if application.Commands.RebuildIndex == nil {
						return ErrMissingRuntime
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
