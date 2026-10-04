package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app"
)

func newDeleteCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name: "delete", Usage: "delete a note document and its index entry by UUID", ArgsUsage: "ID",
		Arguments: []urfavecli.Argument{
			&urfavecli.StringArgs{Name: "id", UsageText: "ID", Min: 1, Max: 1},
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}
			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.DeleteNote(ctx, app.DeleteNoteCommand{ID: command.StringArgs("id")[0]})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(command.Writer, "Deleted note %s.\n", result.ID)
			return err
		},
	}
}
