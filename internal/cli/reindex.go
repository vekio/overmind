package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app/notes"
)

func newReindexCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name: "reindex", Usage: "rebuild the complete index from AsciiDoc note files",
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}
			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.Reindex(ctx, notes.ReindexCommand{})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(command.Writer, "Indexed %d notes.\n", result.Indexed)
			return err
		},
	}
}
