package cli

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app/notes"
)

func newListCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name: "list", Usage: "list indexed notes, optionally filtered by type and tag",
		Flags: []urfavecli.Flag{
			typeFlag(),
			tagFilterFlag(),
			limitFlag(),
			offsetFlag(),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}
			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.ListNotes(ctx, notes.ListQuery{Type: command.String("type"), Tag: command.String("tag"), Limit: command.Int("limit"), Offset: command.Int("offset")})
			if err != nil {
				return err
			}
			if len(result.Notes) == 0 {
				_, err = fmt.Fprintln(command.Writer, "No notes found.")
				return err
			}
			output := tabwriter.NewWriter(command.Writer, 0, 4, 2, ' ', 0)
			if _, err := fmt.Fprintln(output, "ID\tTYPE\tNOTE\tTAGS\tUPDATED"); err != nil {
				return err
			}
			for _, note := range result.Notes {
				if _, err := fmt.Fprintf(output, "%s\t%s\t%s\t%s\t%s\n", note.ID, note.Type, note.Label, strings.Join(note.Tags, ", "), note.UpdatedAt.Local().Format("2006-01-02 15:04")); err != nil {
					return err
				}
			}
			return output.Flush()
		},
	}
}
