package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.casta.me/alberto/overmind/internal/bootstrap"
	overmindcli "git.casta.me/alberto/overmind/internal/cli"
	"git.casta.me/alberto/overmind/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	command, err := overmindcli.New(func(cfg config.Config) (overmindcli.Runtime, error) {
		return bootstrap.New(cfg)
	})
	if err == nil {
		err = command.Run(ctx, os.Args)
	}
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
