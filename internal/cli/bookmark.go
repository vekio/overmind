package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app"
)

func newBookmarkCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "bookmark",
		Usage:     "save a link for later",
		ArgsUsage: "URL",
		Arguments: []urfavecli.Argument{
			&urfavecli.StringArgs{
				Name:      "url",
				UsageText: "URL",
				Min:       1,
				Max:       1,
			},
		},
		Flags: []urfavecli.Flag{
			tagFlag("add `TAG`; repeatable"),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			urlArguments := command.StringArgs("url")
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after URL: %q", command.Args().Slice())
			}

			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.CreateBookmark(ctx, app.CreateBookmarkCommand{URL: urlArguments[0], Tags: command.StringSlice("tag")})
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(command.Writer, "%s\nID: %s\n", result.Bookmark.Summary(), result.Bookmark.ID()); err != nil {
				return fmt.Errorf("write created bookmark summary: %w", err)
			}
			return nil
		},
	}
}
