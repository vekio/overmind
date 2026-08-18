// Package server contains the command-line application for running and
// administering the Overmind HTTP server.
package server

import (
	"context"

	"git.casta.me/alberto/overmind/internal/cli/shared"
	"git.casta.me/alberto/overmind/internal/config"
	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
)

// Runtime contains the application lifecycle and logging capabilities required
// by the server CLI.
type Runtime interface {
	shared.Runtime
	shared.LoggingRuntime
}

type applicationState = shared.State[config.ServerConfig]

// NewCommand creates the complete overmind-server command tree.
func NewCommand(build func(config.ServerConfig) (Runtime, error)) (*urfavecli.Command, error) {
	configFile, defaults, err := config.NewServerFile()
	if err != nil {
		return nil, err
	}
	state := shared.NewState[config.ServerConfig]()

	return &urfavecli.Command{
		Name:  "overmind-server",
		Usage: "serve and administer the Overmind HTTP API",
		Flags: []urfavecli.Flag{
			configlib.NewConfigFlag(configFile),
			shared.DebugFlag("OVERMIND_SERVER_DEBUG"),
		},
		Before: prepareApplication(state, configFile, defaults, build),
		After: func(_ context.Context, _ *urfavecli.Command) error {
			return state.Close()
		},
		Commands: []*urfavecli.Command{
			configlib.NewConfigCommand(configFile, defaults),
			shared.NewIndexCommand(state, "manage the server document index"),
			newServeCommand(state),
		},
	}, nil
}

func prepareApplication(
	state *applicationState,
	configFile *configlib.ConfigFile[config.ServerConfig],
	defaults config.ServerConfig,
	build func(config.ServerConfig) (Runtime, error),
) urfavecli.BeforeFunc {
	return func(ctx context.Context, command *urfavecli.Command) (context.Context, error) {
		if !requiresApplication(command.Args().First()) {
			return ctx, nil
		}
		if build == nil {
			return ctx, shared.ErrMissingRuntimeBuilder
		}

		var cfg config.ServerConfig
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
		return ctx, state.Initialize(cfg, func(cfg config.ServerConfig) (shared.Runtime, error) {
			return build(cfg)
		})
	}
}

func requiresApplication(commandName string) bool {
	switch commandName {
	case "index", "serve":
		return true
	default:
		return false
	}
}
