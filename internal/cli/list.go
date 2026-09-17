package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"

	"git.casta.me/alberto/overmind/internal/app/listdocuments"
	"git.casta.me/alberto/overmind/internal/config"
	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
)

func newListCommand(configFile *configlib.ConfigFile[config.Config]) *urfavecli.Command {
	return &urfavecli.Command{
		Name:    "ls",
		Aliases: []string{"list"},
		Usage:   "list indexed documents",
		Flags: []urfavecli.Flag{
			documentTypeFlag(),
			titleFlag(),
			areaFlag("filter by `AREA`"),
			tagFlag("filter by `TAG`; repeatable"),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) (err error) {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}
			application, err := loadApp(configFile)
			if err != nil {
				return err
			}
			defer func() { err = errors.Join(err, application.Close()) }()

			result, err := application.Queries.ListDocuments.Handle(ctx, listdocuments.ListDocumentsQuery{
				Type:  command.String("type"),
				Title: command.String("title"),
				Area:  command.String("area"),
				Tags:  command.StringSlice("tag"),
			})
			if err != nil {
				return err
			}

			output := bufio.NewWriter(command.Writer)
			for _, document := range result.Documents {
				if _, err := fmt.Fprintf(output, "%s\t%s\t%s\t%s\t%s\n",
					document.ID,
					document.Type,
					document.Area,
					document.Title,
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
