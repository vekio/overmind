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
		Flags: []urfavecli.Flag{areaFlag()},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			application, err := state.get()
			if err != nil {
				return err
			}

			result, err := application.Commands.CreatePage.Handle(ctx, createpage.CreatePageCommand{
				Title: command.StringArgs("title")[0],
				Area:  command.String("area"),
			})
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintln(command.Writer, result.ID); err != nil {
				return fmt.Errorf("write created page id: %w", err)
			}
			return nil
		},
	}
}

func areaFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:    "area",
		Aliases: []string{"a"},
		Usage:   "page area",
		Config:  urfavecli.StringConfig{TrimSpace: true},
	}
}
