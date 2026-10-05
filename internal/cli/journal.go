package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app/journal"
)

func newJournalCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "journal",
		Usage: "create a journal note",
		Flags: []urfavecli.Flag{
			tagFlag(),
			dateFlag(),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}

			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.CreateJournal(ctx, journal.CreateCommand{Date: command.String("date"), Tags: command.StringSlice("tag")})
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(command.Writer, "%s\nID: %s\n", result.Journal.Summary(), result.Journal.ID()); err != nil {
				return fmt.Errorf("write created journal location: %w", err)
			}
			return nil
		},
	}
}
