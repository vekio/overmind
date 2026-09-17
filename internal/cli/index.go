package cli

import (
	"context"
	"errors"
	"fmt"

	"git.casta.me/alberto/overmind/internal/app/rebuildindex"
	"git.casta.me/alberto/overmind/internal/config"
	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
)

func newIndexCommand(configFile *configlib.ConfigFile[config.Config]) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "index",
		Usage: "manage the local document index",
		Commands: []*urfavecli.Command{
			{
				Name:  "rebuild",
				Usage: "rebuild the index from stored AsciiDoc documents",
				Action: func(ctx context.Context, command *urfavecli.Command) (err error) {
					application, err := loadApp(configFile)
					if err != nil {
						return err
					}
					defer func() { err = errors.Join(err, application.Close()) }()

					result, err := application.Commands.RebuildIndex.Handle(ctx, rebuildindex.RebuildIndexCommand{})
					if err != nil {
						return err
					}
					_, err = fmt.Fprintf(command.Writer, "Indexed %d documents\n", result.Documents)
					return err
				},
			},
		},
	}
}
