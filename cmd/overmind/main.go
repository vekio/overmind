package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.casta.me/alberto/overmind/internal/bootstrap"
	"git.casta.me/alberto/overmind/internal/cli"
	"git.casta.me/alberto/overmind/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cli.NewCommand(func(cfg config.Config) (cli.Runtime, error) {
		return bootstrap.NewContainer(cfg)
	}).Run(ctx, os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
