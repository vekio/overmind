package cli

import (
	"context"
	"errors"
	"fmt"

	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/config"
	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
)

func newPageCommand(configFile *configlib.ConfigFile[config.Config]) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "page",
		Usage:     "create an AsciiDoc page",
		ArgsUsage: "TITLE",
		Arguments: []urfavecli.Argument{
			&urfavecli.StringArgs{
				Name:      "title",
				UsageText: "TITLE",
				Min:       1,
				Max:       1,
			},
		},
		Flags: []urfavecli.Flag{
			areaFlag("organize under `AREA`"),
			tagFlag("add `TAG`; repeatable"),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) (err error) {
			titleArguments := command.StringArgs("title")
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after title: %q", command.Args().Slice())
			}
			application, err := loadApp(configFile)
			if err != nil {
				return err
			}
			defer func() { err = errors.Join(err, application.Close()) }()

			result, err := application.Commands.CreatePage.Handle(ctx, createpage.CreatePageCommand{
				Title: titleArguments[0],
				Area:  command.String("area"),
				Tags:  command.StringSlice("tag"),
			})
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintln(command.Writer, result.ID); err != nil {
				return fmt.Errorf("write created page ID: %w", err)
			}
			return nil
		},
	}
}
