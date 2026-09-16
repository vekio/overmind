// Package cli defines the Overmind command-line application.
package cli

import (
	"context"

	"git.casta.me/alberto/overmind/internal/config"
	asciidoccli "git.casta.me/alberto/overmind/pkg/asciidoc/cli"
	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
	configurfave "github.com/vekio/config/urfave"
)

// New creates the complete Overmind command tree.
func New(build func(config.Config) (Runtime, error)) (*urfavecli.Command, error) {
	configFile, defaults, err := config.NewFile()
	if err != nil {
		return nil, err
	}
	state := newApplicationState()

	return &urfavecli.Command{
		Name:                  "overmind",
		Usage:                 "manage the overmind knowledge base",
		EnableShellCompletion: true,
		Flags: []urfavecli.Flag{
			configurfave.NewConfigFlag(configFile),
		},
		Before: prepareApplication(state, configFile, defaults, build),
		After: func(_ context.Context, _ *urfavecli.Command) error {
			return state.Close()
		},
		Commands: []*urfavecli.Command{
			asciidoccli.Command(),
			configurfave.NewConfigCommand(configFile, defaults),
			newEditCommand(state),
			newIndexCommand(state),
			newListCommand(state),
			newPageCommand(state),
			newShowCommand(state),
		},
	}, nil
}

func prepareApplication(
	state *applicationState,
	configFile *configlib.ConfigFile[config.Config],
	defaults config.Config,
	build func(config.Config) (Runtime, error),
) urfavecli.BeforeFunc {
	return func(ctx context.Context, command *urfavecli.Command) (context.Context, error) {
		if !requiresApplication(command.Args().First()) {
			return ctx, nil
		}

		var cfg config.Config
		var err error
		if command.IsSet("config") {
			cfg, err = configFile.Load()
		} else {
			cfg, err = configFile.LoadOrCreate(defaults)
		}
		if err != nil {
			return ctx, err
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
