package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
	"github.com/vekio/overmind/internal/bootstrap"
	"github.com/vekio/overmind/internal/cli"
	appconfig "github.com/vekio/overmind/internal/config"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (runErr error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	configFile, err := appconfig.NewFile()
	if err != nil {
		return err
	}
	runtime := bootstrap.New(configFile)
	defer func() {
		if err := runtime.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("close application: %w", err))
		}
	}()

	return newCommand(configFile, runtime).Run(ctx, os.Args)
}

func newCommand(configFile *configlib.ConfigFile[appconfig.Settings], runtime *bootstrap.Runtime) *urfavecli.Command {
	return cli.New(configFile, func(ctx context.Context) (cli.Client, error) {
		application, err := runtime.Application(ctx)
		if err != nil {
			return nil, err
		}
		return cli.NewLocalClient(application), nil
	})
}
