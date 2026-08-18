package shared

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/app/rebuildindex"
	urfavecli "github.com/urfave/cli/v3"
)

// NewIndexCommand creates the index administration command shared by both
// binaries.
func NewIndexCommand[T any](state *State[T], usage string) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "index",
		Usage: usage,
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
