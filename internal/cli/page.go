package cli

import (
	"context"
	"fmt"
	"time"

	"git.casta.me/alberto/overmind/internal/id"
	"git.casta.me/alberto/overmind/internal/renderer"
	urfavecli "github.com/urfave/cli/v3"
)

func newPageCommand(documentRenderer *renderer.Renderer) *urfavecli.Command {
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
			areaFlag("organize under `AREA`"),
			tagFlag("add `TAG`; repeatable"),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			titleArguments := command.StringArgs("title")
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after title: %q", command.Args().Slice())
			}

			now := time.Now()
			content, err := documentRenderer.Render(ctx, renderer.PageTemplate, renderer.Page{
				ID:        id.New().String(),
				Title:     titleArguments[0],
				Area:      command.String("area"),
				Tags:      command.StringSlice("tag"),
				CreatedAt: now,
				UpdatedAt: now,
			})
			if err != nil {
				return err
			}

			if _, err := command.Writer.Write(content); err != nil {
				return fmt.Errorf("write rendered page: %w", err)
			}
			return nil
		},
	}
}
