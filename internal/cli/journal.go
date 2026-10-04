package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app"
)

func newJournalCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "journal",
		Usage: "create a journal note",
		Flags: []urfavecli.Flag{
			tagFlag("add `TAG`; repeatable"),
			&urfavecli.StringFlag{Name: "date", Usage: "calendar date YYYY-MM-DD; defaults to today"},
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}

			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.CreateJournal(ctx, app.CreateJournalCommand{Date: command.String("date"), Tags: command.StringSlice("tag")})
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
