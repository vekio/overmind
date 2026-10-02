package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
	configurfave "github.com/vekio/config/urfave"
	appconfig "github.com/vekio/overmind/internal/config"
	"github.com/vekio/overmind/internal/tui"
)

func New(configFile *configlib.ConfigFile[appconfig.Settings], newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:                  "overmind",
		Version:               buildVersion(),
		Usage:                 "manage the overmind knowledge base",
		Description:           "Run without a command to open the terminal interface.",
		EnableShellCompletion: true,
		Flags:                 []urfavecli.Flag{configurfave.NewConfigFlag(configFile)},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}
			client, err := newClient(ctx)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("configuration missing; run 'overmind setup': %w", err)
				}
				return err
			}
			return tui.Run(ctx, client)
		},
		Commands: []*urfavecli.Command{
			newBookmarkCommand(newClient),
			newCaptureCommand(newClient),
			newJournalCommand(newClient),
			newPageCommand(newClient),
			newPersonCommand(newClient),
			newRebuildCommand(newClient),
			configurfave.NewConfigCommand(configFile),
			NewSetupCommand(configFile),
		},
	}
}
