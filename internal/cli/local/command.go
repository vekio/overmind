// Package local contains the command-line application for working directly
// with a local Overmind vault.
package local

import (
	"context"

	"git.casta.me/alberto/overmind/internal/cli/shared"
	"git.casta.me/alberto/overmind/internal/config"
	asciidoccli "git.casta.me/alberto/overmind/pkg/asciidoc/cli"
	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
)

// Runtime is the application runtime required by the local CLI.
type Runtime = shared.Runtime

type applicationState = shared.State[config.CLIConfig]

// NewCommand creates the complete overmind command tree.
func NewCommand(build func(config.CLIConfig) (Runtime, error)) (*urfavecli.Command, error) {
	configFile, defaults, err := config.NewCLIFile()
	if err != nil {
		return nil, err
	}
	state := shared.NewState[config.CLIConfig]()

	return &urfavecli.Command{
		Name:                  "overmind",
		Usage:                 "manage the overmind knowledge base",
		EnableShellCompletion: true,
		Flags: []urfavecli.Flag{
			configlib.NewConfigFlag(configFile),
			shared.DebugFlag("OVERMIND_DEBUG"),
		},
		Before: prepareApplication(state, configFile, defaults, build),
		After: func(_ context.Context, _ *urfavecli.Command) error {
			return state.Close()
		},
		Commands: []*urfavecli.Command{
			asciidoccli.Command(),
			configlib.NewConfigCommand(configFile, defaults),
			newEditCommand(state),
			shared.NewIndexCommand(state, "manage the local document index"),
			newListCommand(state),
			newPageCommand(state),
			newShowCommand(state),
		},
	}, nil
}

func prepareApplication(
	state *applicationState,
	configFile *configlib.ConfigFile[config.CLIConfig],
	defaults config.CLIConfig,
	build func(config.CLIConfig) (Runtime, error),
) urfavecli.BeforeFunc {
	return func(ctx context.Context, command *urfavecli.Command) (context.Context, error) {
		if !requiresApplication(command.Args().First()) {
			return ctx, nil
		}

		var cfg config.CLIConfig
		var err error
		if command.IsSet("config") {
			cfg, err = configFile.Load()
		} else {
			cfg, err = configFile.LoadOrCreate(defaults)
		}
		if err != nil {
			return ctx, err
		}
		if command.Bool("debug") {
			cfg.Logging.Level = "debug"
		}
		return ctx, state.Initialize(cfg, build)
	}
}

func requiresApplication(commandName string) bool {
	switch commandName {
	case "edit", "index", "list", "ls", "page", "show":
		return true
	default:
		return false
	}
}
