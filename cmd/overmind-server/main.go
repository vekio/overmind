package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.casta.me/alberto/overmind/internal/bootstrap"
	servercli "git.casta.me/alberto/overmind/internal/cli/server"
	"git.casta.me/alberto/overmind/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	command, err := servercli.NewCommand(func(cfg config.ServerConfig) (servercli.Runtime, error) {
		return bootstrap.NewServerContainer(cfg)
	})
	if err == nil {
		err = command.Run(ctx, os.Args)
	}
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
