package cli

import (
	"context"
	"fmt"
	"time"

	"git.casta.me/alberto/overmind/internal/id"
	"git.casta.me/alberto/overmind/internal/renderer"
	"git.casta.me/alberto/overmind/internal/storage"
	urfavecli "github.com/urfave/cli/v3"
)

func newJournalCommand(documentRenderer *renderer.Renderer, noteWriter *storage.FileWriter) *urfavecli.Command {
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
			noteID := id.New()
			content, err := documentRenderer.Render(ctx, renderer.JournalTemplate, renderer.Journal{
				ID:        noteID.String(),
				Date:      now.Format(time.DateOnly),
				Tags:      command.StringSlice("tag"),
				CreatedAt: now,
				UpdatedAt: now,
			})
			if err != nil {
				return err
			}

			path, err := noteWriter.Write(ctx, noteID, content)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(command.Writer, path); err != nil {
				return fmt.Errorf("write created journal path: %w", err)
			}
			return nil
		},
	}
}
