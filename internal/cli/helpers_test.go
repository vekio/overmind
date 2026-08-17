package cli

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
)

type runtimeStub struct {
	application *app.Application
}

func (runtime runtimeStub) Application() *app.Application { return runtime.application }

func (runtime runtimeStub) Logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (runtime runtimeStub) Close() error { return nil }

func writeLocalConfig(t *testing.T) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(configPath, []byte("vault:\n  localRoot: ./vault\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return configPath
}
