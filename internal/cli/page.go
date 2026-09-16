package cli

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/app/createpage"
	urfavecli "github.com/urfave/cli/v3"
)

func newPageCommand(state *applicationState) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "page",
		Usage:     "create an AsciiDoc page",
		ArgsUsage: "TITLE",
		Arguments: []urfavecli.Argument{
			&urfavecli.StringArgs{
				Name:      "title",
				UsageText: "TITLE",
				Min:       1,
				Max:       1,
			},
		},
		Flags: []urfavecli.Flag{
			areaFlag(),
			tagFlag("add `TAG`; repeatable"),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			titleArguments := command.StringArgs("title")
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after title: %q", command.Args().Slice())
			}
			application, err := state.Application()
			if err != nil {
				return err
			}

			result, err := application.Commands.CreatePage.Handle(ctx, createpage.CreatePageCommand{
				Title: titleArguments[0],
				Area:  command.String("area"),
				Tags:  command.StringSlice("tag"),
			})
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintln(command.Writer, result.Path); err != nil {
				return fmt.Errorf("write created page path: %w", err)
			}
			return nil
		},
	}
}

func areaFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:    "area",
		Aliases: []string{"a"},
		Usage:   "organize under `AREA`",
		Config:  urfavecli.StringConfig{TrimSpace: true},
	}
}
