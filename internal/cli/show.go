package cli

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/app/getdocument"
	urfavecli "github.com/urfave/cli/v3"
)

func newShowCommand(state *applicationState) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "show",
		Usage:     "write a document to standard output",
		ArgsUsage: "[PATH]",
		Arguments: []urfavecli.Argument{
			&urfavecli.StringArgs{Name: "path", UsageText: "PATH", Min: 0, Max: 1},
		},
		ShellComplete: func(ctx context.Context, command *urfavecli.Command) {
			completeDocumentPaths(ctx, command, state)
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after path: %q", command.Args().Slice())
			}
			path, err := documentPath(command, "path")
			if err != nil {
				return err
			}
			application, err := state.Application()
			if err != nil {
				return err
			}
			result, err := application.Queries.GetDocument.Handle(ctx, getdocument.GetDocumentQuery{
				Path: path,
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
