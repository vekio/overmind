package cli

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"git.casta.me/alberto/overmind/internal/app/listdocuments"
	urfavecli "github.com/urfave/cli/v3"
)

func newListCommand(state *applicationState) *urfavecli.Command {
	return &urfavecli.Command{
		Name:    "ls",
		Aliases: []string{"list"},
		Usage:   "list indexed documents",
		Flags:   []urfavecli.Flag{documentTypeFlag(), tagFlag()},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}
			application, err := state.get()
			if err != nil {
				return err
			}

			result, err := application.Queries.ListDocuments.Handle(ctx, listdocuments.ListDocumentsQuery{
				Type: command.String("type"),
				Tags: command.StringSlice("tag"),
			})
			if err != nil {
				return err
			}

			output := bufio.NewWriter(command.Writer)
			for _, document := range result.Documents {
				if _, err := fmt.Fprintf(output, "%s\t%s\t%s\n",
					document.Path,
					document.Type,
					strings.Join(document.Tags, ","),
				); err != nil {
					return fmt.Errorf("write document list: %w", err)
				}
			}
			if err := output.Flush(); err != nil {
				return fmt.Errorf("write document list: %w", err)
			}
			return nil
		},
	}
}
