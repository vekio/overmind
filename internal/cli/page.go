package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app/page"
)

func newPageCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "page",
		Usage:     "create a page note",
		ArgsUsage: "TITLE",
		Arguments: titleArgument(),
		Flags: []urfavecli.Flag{
			areaFlag(),
			tagFlag(),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			titleArguments := command.StringArgs("title")
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after title: %q", command.Args().Slice())
			}

			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.CreatePage(ctx, page.CreateCommand{Title: titleArguments[0], Area: command.String("area"), Tags: command.StringSlice("tag")})
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(command.Writer, "%s\nID: %s\n", result.Page.Summary(), result.Page.ID()); err != nil {
				return fmt.Errorf("write created page location: %w", err)
			}
			return nil
		},
	}
}
