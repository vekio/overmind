package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
	appconfig "github.com/vekio/overmind/internal/config"
	"github.com/vekio/overmind/internal/tui"
)

type settingsPrompt func(context.Context, appconfig.Settings) (appconfig.Settings, error)

// NewSetupCommand creates the interactive initial configuration command.
func NewSetupCommand(configFile *configlib.ConfigFile[appconfig.Settings]) *urfavecli.Command {
	return newSetupCommand(configFile, func(ctx context.Context, defaults appconfig.Settings) (appconfig.Settings, error) {
		return tui.RunSetup(ctx, defaults, configFile.Path())
	})
}

func newSetupCommand(configFile *configlib.ConfigFile[appconfig.Settings], prompt settingsPrompt) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "setup",
		Usage: "create the initial configuration interactively",
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if _, err := os.Stat(configFile.Path()); err == nil {
				return fmt.Errorf("configuration already exists at %q", configFile.Path())
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("inspect configuration file: %w", err)
			}

			settings, err := prompt(ctx, configFile.Defaults())
			if err != nil {
				return fmt.Errorf("setup: %w", err)
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			settings.Mode = appconfig.Mode(strings.TrimSpace(string(settings.Mode)))
			if settings.Mode == "" {
				settings.Mode = appconfig.ModeLocal
			}
			settings.VaultPath, err = appconfig.NormalizeVaultPath(settings.VaultPath)
			if err != nil {
				return err
			}
			if err := settings.Validate(); err != nil {
				return err
			}
			if err := configFile.Create(settings); err != nil {
				return fmt.Errorf("create configuration: %w", err)
			}
			_, err = fmt.Fprintln(command.Writer, configFile.Path())
			return err
		},
	}
}
