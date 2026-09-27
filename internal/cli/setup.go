package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/huh/v2"
	appconfig "git.casta.me/alberto/overmind/internal/config"
	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
)

type vaultPrompt func(context.Context, string) (string, error)

// NewSetupCommand creates the interactive configuration command.
func NewSetupCommand(configFile *configlib.ConfigFile[appconfig.Settings]) *urfavecli.Command {
	return newSetupCommand(configFile, promptVault)
}

func newSetupCommand(configFile *configlib.ConfigFile[appconfig.Settings], prompt vaultPrompt) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "setup",
		Usage: "configure Overmind interactively",
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}
			if _, err := os.Stat(configFile.Path()); err == nil {
				return fmt.Errorf("configuration already exists at %q", configFile.Path())
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("inspect configuration file: %w", err)
			}

			vault, err := prompt(ctx, configFile.Defaults().VaultPath)
			if err != nil {
				return fmt.Errorf("setup: %w", err)
			}
			vault, err = normalizeVaultPath(vault)
			if err != nil {
				return err
			}
			settings := appconfig.Settings{Mode: appconfig.ModeLocal, VaultPath: vault}
			if err := configFile.Create(settings); err != nil {
				return fmt.Errorf("create configuration: %w", err)
			}
			_, err = fmt.Fprintln(command.Writer, configFile.Path())
			return err
		},
	}
}

func promptVault(ctx context.Context, defaultPath string) (string, error) {
	vault := defaultPath
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Vault directory").
			Description("Directory where Overmind stores notes and its index").
			Value(&vault).
			Validate(func(value string) error {
				_, err := normalizeVaultPath(value)
				return err
			}),
	))
	if err := form.RunWithContext(ctx); err != nil {
		return "", err
	}
	return vault, nil
}

func normalizeVaultPath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("vault directory is required")
	}
	if value == "~" || strings.HasPrefix(value, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		if value == "~" {
			value = home
		} else {
			value = filepath.Join(home, strings.TrimPrefix(value, "~/"))
		}
	} else if strings.HasPrefix(value, "~") {
		return "", fmt.Errorf("vault directory cannot use another user's home")
	}
	path, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("resolve vault directory: %w", err)
	}
	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("vault path %q is not a directory", path)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect vault directory: %w", err)
	}
	return path, nil
}
