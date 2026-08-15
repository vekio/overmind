package cli

import (
	"context"
	"errors"
	"log/slog"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/config"
	urfavecli "github.com/urfave/cli/v3"
)

var (
	errMissingRuntime        = errors.New("missing application runtime")
	errMissingRuntimeBuilder = errors.New("missing application runtime builder")
)

// Runtime is the configured application runtime needed by the CLI. Bootstrap's
// container implements it without exposing infrastructure dependencies here.
type Runtime interface {
	Application() *app.Application
	Logger() *slog.Logger
	Close() error
}

func (state *applicationState) close() error {
	if state == nil || state.runtime == nil {
		return nil
	}
	return state.runtime.Close()
}

type applicationState struct {
	runtime Runtime
	config  config.Config
}

func (state *applicationState) get() (*app.Application, error) {
	if state == nil || state.runtime == nil {
		return nil, errMissingRuntime
	}
	application := state.runtime.Application()
	if application == nil {
		return nil, errMissingRuntime
	}
	return application, nil
}

func prepareApplication(
	state *applicationState,
	build func(config.Config) (Runtime, error),
) urfavecli.BeforeFunc {
	return func(ctx context.Context, command *urfavecli.Command) (context.Context, error) {
		if !requiresApplication(command) {
			return ctx, nil
		}
		if build == nil {
			return ctx, errMissingRuntimeBuilder
		}

		cfg, err := loadConfig(command.String("config"))
		if err != nil {
			return ctx, err
		}
		runtime, err := build(cfg)
		if err != nil {
			return ctx, err
		}
		state.runtime = runtime
		state.config = cfg
		return ctx, nil
	}
}

func requiresApplication(command *urfavecli.Command) bool {
	switch command.Args().First() {
	case "document", "index", "page", "serve":
		return true
	default:
		return false
	}
}

func loadConfig(path string) (config.Config, error) {
	if path == "" {
		return config.Load()
	}
	return config.LoadFrom(path)
}
