package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.casta.me/alberto/overmind/internal/bootstrap"
	"git.casta.me/alberto/overmind/internal/cli"
	appconfig "git.casta.me/alberto/overmind/internal/config"
	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
	configurfave "github.com/vekio/config/urfave"
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
	command := cli.New(func(ctx context.Context) (cli.Client, error) {
		application, err := runtime.Application(ctx)
		if err != nil {
			return nil, err
		}
		return cli.NewLocalClient(application), nil
	})
	command.Flags = append(command.Flags, configurfave.NewConfigFlag(configFile))
	command.Commands = append(command.Commands, configurfave.NewConfigCommand(configFile))
	return command
}
