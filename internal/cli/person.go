package cli

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/domain"
	urfavecli "github.com/urfave/cli/v3"
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

			name, err := domain.NewTitle(nameArguments[0])
			if err != nil {
				return err
			}

			values := make([]domain.Group, 0)
			for _, value := range command.StringSlice("group") {
				group, err := domain.NewGroup(value)
				if err != nil {
					return err
				}
				values = append(values, group)
			}
			groups, err := domain.NewGroups(values...)
			if err != nil {
				return err
			}

			tags, err := parseTags(command.StringSlice("tag"))
			if err != nil {
				return err
			}

			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.CreatePerson(ctx, name, groups, tags)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(command.Writer, result.Path); err != nil {
				return fmt.Errorf("write created person location: %w", err)
			}
			return nil
		},
	}
}
