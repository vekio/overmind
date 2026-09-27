package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
)

func newRebuildCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "rebuild",
		Usage: "rebuild the index from AsciiDoc notes",
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}
			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			count, err := client.RebuildIndex(ctx)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(command.Writer, "indexed %d note(s)\n", count)
			return err
		},
	}
}
