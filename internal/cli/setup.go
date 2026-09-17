package cli

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/config"
	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
)

func newSetupCommand(
	configFile *configlib.ConfigFile[config.Config],
	defaults config.Config,
) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "setup",
		Usage: "initialize the local configuration",
		Action: func(_ context.Context, command *urfavecli.Command) error {
			if _, err := configFile.LoadOrCreate(defaults); err != nil {
				return fmt.Errorf("setup configuration: %w", err)
			}
			_, err := fmt.Fprintf(command.Root().Writer, "Configuration: %s\n", configFile.Path())
			return err
		},
	}
}
