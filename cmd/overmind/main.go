package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	overmindcli "git.casta.me/alberto/overmind/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	command, err := overmindcli.New()
	if err == nil {
		err = command.Run(ctx, os.Args)
	}
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
