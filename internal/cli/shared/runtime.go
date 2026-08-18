// Package shared contains command-line infrastructure shared by the local
// client and server administration binaries.
package shared

import (
	"errors"
	"log/slog"

	"git.casta.me/alberto/overmind/internal/app"
)

var (
	ErrMissingRuntime        = errors.New("missing application runtime")
	ErrMissingRuntimeBuilder = errors.New("missing application runtime builder")
	ErrMissingLogger         = errors.New("runtime does not provide a logger")
)

// Runtime is the application lifecycle required by commands that execute use
// cases. Bootstrap containers implement it without exposing infrastructure
// details to the command-line adapter.
type Runtime interface {
	Application() *app.Application
	Close() error
}

// LoggingRuntime is an optional runtime capability used by commands that need
// direct access to the process logger, such as the HTTP server.
type LoggingRuntime interface {
	Logger() *slog.Logger
}

// State holds the lazily-created runtime and the configuration used to build
// it for one command invocation.
type State[T any] struct {
	runtime Runtime
	config  T
}

// NewState creates an empty command invocation state.
func NewState[T any]() *State[T] { return &State[T]{} }

// Initialize builds and stores the runtime for cfg.
func (state *State[T]) Initialize(cfg T, build func(T) (Runtime, error)) error {
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
	state.config = cfg
	return nil
}

// Application returns the configured use-case registry.
func (state *State[T]) Application() (*app.Application, error) {
	if state == nil || state.runtime == nil {
		return nil, ErrMissingRuntime
	}
	application := state.runtime.Application()
	if application == nil {
		return nil, ErrMissingRuntime
	}
	return application, nil
}

// Logger returns the logger owned by the configured runtime.
func (state *State[T]) Logger() (*slog.Logger, error) {
	if state == nil || state.runtime == nil {
		return nil, ErrMissingRuntime
	}
	loggingRuntime, ok := state.runtime.(LoggingRuntime)
	if !ok {
		return nil, ErrMissingLogger
	}
	logger := loggingRuntime.Logger()
	if logger == nil {
		return nil, ErrMissingLogger
	}
	return logger, nil
}

// Config returns the configuration used to build the runtime.
func (state *State[T]) Config() T {
	if state == nil {
		var zero T
		return zero
	}
	return state.config
}

// Close releases the configured runtime, if one was created.
func (state *State[T]) Close() error {
	if state == nil || state.runtime == nil {
		return nil
	}
	return state.runtime.Close()
}
