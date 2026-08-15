package cli

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/app/getdocument"
	urfavecli "github.com/urfave/cli/v3"
)

func newDocumentCommand(state *applicationState) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "document",
		Usage:     "get a document by ID",
		ArgsUsage: "ID",
		Arguments: []urfavecli.Argument{
			&urfavecli.StringArgs{Name: "id", UsageText: "ID", Min: 1, Max: 1},
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			application, err := state.get()
			if err != nil {
				return err
			}
			if application.Queries.GetDocument == nil {
				return errMissingRuntime
			}
			result, err := application.Queries.GetDocument.Handle(ctx, getdocument.GetDocumentQuery{
				ID: command.StringArgs("id")[0],
			})
			if err != nil {
				return err
			}
			if _, err := command.Writer.Write(result.Content); err != nil {
				return fmt.Errorf("write document: %w", err)
			}
			return nil
		},
	}
}
