package cli

import (
	"context"
	"fmt"
	"time"

	"git.casta.me/alberto/overmind/internal/id"
	"git.casta.me/alberto/overmind/internal/renderer"
	urfavecli "github.com/urfave/cli/v3"
)

func newJournalCommand(documentRenderer *renderer.Renderer) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "journal",
		Usage: "create a journal note",
		Flags: []urfavecli.Flag{
			tagFlag("add `TAG`; repeatable"),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}

			now := time.Now()
			content, err := documentRenderer.Render(ctx, renderer.JournalTemplate, renderer.Journal{
				ID:        id.New().String(),
				Date:      now.Format(time.DateOnly),
				Tags:      command.StringSlice("tag"),
				CreatedAt: now,
				UpdatedAt: now,
			})
			if err != nil {
				return err
			}

			if _, err := command.Writer.Write(content); err != nil {
				return fmt.Errorf("write rendered journal: %w", err)
			}
			return nil
		},
	}
}
