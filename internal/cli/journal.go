package cli

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/app"
	urfavecli "github.com/urfave/cli/v3"
)

func newJournalCommand(application *app.App) *urfavecli.Command {
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

			tags, err := parseTags(command.StringSlice("tag"))
			if err != nil {
				return err
			}

			path, err := application.CreateJournal(ctx, tags)
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
