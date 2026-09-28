package cli

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/domain"
	urfavecli "github.com/urfave/cli/v3"
)

func newPageCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "page",
		Usage:     "create a page note",
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
			areaFlag("organize under `AREA`"),
			tagFlag("add `TAG`; repeatable"),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			titleArguments := command.StringArgs("title")
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after title: %q", command.Args().Slice())
			}

			title, err := domain.NewTitle(titleArguments[0])
			if err != nil {
				return err
			}

			var area domain.Area
			if value := command.String("area"); value != "" {
				area, err = domain.NewArea(value)
				if err != nil {
					return err
				}
			}

			tags, err := parseTags(command.StringSlice("tag"))
			if err != nil {
				return err
			}

			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.CreatePage(ctx, title, area, tags)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(command.Writer, result.Path); err != nil {
				return fmt.Errorf("write created page location: %w", err)
			}
			return nil
		},
	}
}
