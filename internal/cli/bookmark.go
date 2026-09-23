package cli

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
	urfavecli "github.com/urfave/cli/v3"
)

func newBookmarkCommand(application *app.App) *urfavecli.Command {
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

			bookmarkURL, err := domain.NewURL(urlArguments[0])
			if err != nil {
				return err
			}
			tags, err := parseTags(command.StringSlice("tag"))
			if err != nil {
				return err
			}

			path, err := application.CreateBookmark(ctx, bookmarkURL, tags)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(command.Writer, path); err != nil {
				return fmt.Errorf("write created bookmark path: %w", err)
			}
			return nil
		},
	}
}
