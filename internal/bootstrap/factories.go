package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"git.casta.me/alberto/overmind/internal/config"
	"git.casta.me/alberto/overmind/internal/infrastructure/gotemplate"
	"git.casta.me/alberto/overmind/internal/infrastructure/localfs"
	"git.casta.me/alberto/overmind/internal/infrastructure/sqliteindex"
	"git.casta.me/alberto/overmind/internal/infrastructure/systemclock"
	"git.casta.me/alberto/overmind/internal/infrastructure/uuidgenerator"
	"git.casta.me/alberto/overmind/internal/ports"
)

func newClock() ports.Clock {
	return systemclock.New()
}

func newIDGenerator() ports.IDGenerator {
	return uuidgenerator.New()
}

func newLogger(cfg config.Logging) (*slog.Logger, error) {
	var level slog.Level

	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return nil, fmt.Errorf("unsupported log level %q", cfg.Level)
	}

	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})), nil
}

func newBlobStore(cfg config.Vault) (ports.BlobStore, error) {
	switch cfg.Driver {
	case "local":
		return localfs.New(cfg.RootPath), nil
	default:
		return nil, fmt.Errorf("unsupported vault driver %q", cfg.Driver)
	}
}

func newDocumentIndex(cfg config.Index) (*sqliteindex.Store, error) {
	switch cfg.Driver {
	case "sqlite":
		return sqliteindex.New(context.Background(), cfg.Path)
	default:
		return nil, fmt.Errorf("unsupported index driver %q", cfg.Driver)
	}
}

func newRenderer() (ports.Renderer, error) {
	return gotemplate.New()
}
