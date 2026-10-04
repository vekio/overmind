package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app"
)

func newPersonCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "person",
		Usage:     "create a person note",
		ArgsUsage: "NAME",
		Arguments: []urfavecli.Argument{
			&urfavecli.StringArgs{
				Name:      "name",
				UsageText: "NAME",
				Min:       1,
				Max:       1,
			},
		},
		Flags: []urfavecli.Flag{
			&urfavecli.StringSliceFlag{Name: "group", Usage: "add `GROUP`; repeatable"},
			tagFlag("add `TAG`; repeatable"),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			nameArguments := command.StringArgs("name")
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after name: %q", command.Args().Slice())
			}

			input := app.CreatePersonCommand{Name: nameArguments[0], Groups: command.StringSlice("group"), Tags: command.StringSlice("tag")}
			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.CreatePerson(ctx, input)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(command.Writer, "%s\nID: %s\n", result.Person.Summary(), result.Person.ID()); err != nil {
				return fmt.Errorf("write created person summary: %w", err)
			}
			return nil
		},
	}
}
