// Package bootstrap assembles the application and its infrastructure.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	configlib "github.com/vekio/config"
	"github.com/vekio/overmind/internal/app"
	appconfig "github.com/vekio/overmind/internal/config"
)

// Runtime loads and owns the configured application services.
type Runtime struct {
	configFile  *configlib.ConfigFile[appconfig.Settings]
	application *app.Application
	closers     []io.Closer
}

// New creates a lazily configured runtime.
func New(configFile *configlib.ConfigFile[appconfig.Settings]) *Runtime {
	return &Runtime{configFile: configFile}
}

// Application returns the configured application, creating it when first needed.
func (runtime *Runtime) Application(ctx context.Context) (*app.Application, error) {
	if runtime.application != nil {
		return runtime.application, nil
	}

	settings, err := runtime.configFile.Load()
	if err != nil {
		return nil, err
	}
	application, closers, err := buildApplication(ctx, settings)
	if err != nil {
		return nil, err
	}

	runtime.application = application
	runtime.closers = closers
	return runtime.application, nil
}

// Close releases the configured application resources.
func (runtime *Runtime) Close() error {
	var closeErr error
	for _, closer := range slices.Backward(runtime.closers) {
		if err := closer.Close(); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("close %T: %w", closer, err))
		}
	}
	runtime.closers = nil
	return closeErr
}
