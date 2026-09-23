package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/cli"
	"git.casta.me/alberto/overmind/internal/infra/sqliteindex"
	"git.casta.me/alberto/overmind/internal/renderer"
	"git.casta.me/alberto/overmind/internal/storage"
)

const notesRoot = ".overmind/notes"
const indexPath = ".overmind/index.db"

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (runErr error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	documentRenderer, err := renderer.New()
	if err != nil {
		return err
	}
	noteWriter := storage.NewFileWriter(notesRoot)
	noteIndex, err := sqliteindex.New(ctx, indexPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := noteIndex.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("close SQLite index: %w", err))
		}
	}()

	application := app.New(documentRenderer, noteWriter, noteIndex)
	return cli.New(application).Run(ctx, os.Args)
}
