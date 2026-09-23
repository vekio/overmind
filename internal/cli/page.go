package cli

import (
	"context"
	"fmt"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/id"
	"git.casta.me/alberto/overmind/internal/renderer"
	"git.casta.me/alberto/overmind/internal/storage"
	urfavecli "github.com/urfave/cli/v3"
)

func newPageCommand(documentRenderer *renderer.Renderer, noteWriter *storage.FileWriter) *urfavecli.Command {
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

			now := time.Now()
			noteID := id.New()
			page, err := domain.NewPage(noteID, title, area, tags, now)
			if err != nil {
				return err
			}

			content, err := documentRenderer.Render(ctx, renderer.PageTemplate, page)
			if err != nil {
				return err
			}

			path, err := noteWriter.Write(ctx, noteID, content)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(command.Writer, path); err != nil {
				return fmt.Errorf("write created page path: %w", err)
			}
			return nil
		},
	}
}
