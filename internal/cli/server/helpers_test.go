package server

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/config"
	urfavecli "github.com/urfave/cli/v3"
)

type runtimeStub struct {
	application *app.Application
}

func (runtime runtimeStub) Application() *app.Application { return runtime.application }
func (runtime runtimeStub) Logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
func (runtime runtimeStub) Close() error { return nil }

func newTestCommand(
	t *testing.T,
	build func(config.ServerConfig) (Runtime, error),
) *urfavecli.Command {
	t.Helper()
	command, err := NewCommand(build)
	if err != nil {
		t.Fatalf("NewCommand() error = %v", err)
	}
	return command
}

func writeServerConfig(t *testing.T) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(configPath, []byte(`vault:
  driver: local
  localRoot: ./vault
logging:
  level: info
http:
  address: 127.0.0.1:8080
index:
  driver: sqlite
  path: ./index.db
`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return configPath
}
