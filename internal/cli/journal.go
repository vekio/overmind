package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
)

func newJournalCommand(newClient ClientFactory) *urfavecli.Command {
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

			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.CreateJournal(ctx, tags)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(command.Writer, result.Path); err != nil {
				return fmt.Errorf("write created journal location: %w", err)
			}
			return nil
		},
	}
}
