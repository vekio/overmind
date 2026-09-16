package cli

import (
	"errors"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/config"
)

var (
	ErrMissingRuntime        = errors.New("missing application runtime")
	ErrMissingRuntimeBuilder = errors.New("missing application runtime builder")
)

// Runtime is the application lifecycle required by commands that execute use
// cases. Bootstrap runtimes implement it without exposing infrastructure
// details to the command-line adapter.
type Runtime interface {
	Application() *app.Application
	Close() error
}

// applicationState holds the lazily-created runtime for one command invocation.
type applicationState struct {
	runtime Runtime
}

func newApplicationState() *applicationState { return &applicationState{} }

// Initialize builds and stores the runtime for cfg.
func (state *applicationState) Initialize(
	cfg config.Config,
	build func(config.Config) (Runtime, error),
) error {
	if build == nil {
		return ErrMissingRuntimeBuilder
	}
	runtime, err := build(cfg)
	if err != nil {
		return err
	}
	if runtime == nil {
		return ErrMissingRuntime
	}
	state.runtime = runtime
	return nil
}

// Application returns the configured use-case registry.
func (state *applicationState) Application() (*app.Application, error) {
	if state == nil || state.runtime == nil {
		return nil, ErrMissingRuntime
	}
	application := state.runtime.Application()
	if application == nil {
		return nil, ErrMissingRuntime
	}
	return application, nil
}

// Close releases the configured runtime, if one was created.
func (state *applicationState) Close() error {
	if state == nil || state.runtime == nil {
		return nil
	}
	return state.runtime.Close()
}
