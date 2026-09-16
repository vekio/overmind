package cli

import (
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

func (runtime runtimeStub) Close() error { return nil }

func newTestApplicationState(runtime Runtime) *applicationState {
	state := newApplicationState()
	err := state.Initialize(config.Config{}, func(config.Config) (Runtime, error) {
		return runtime, nil
	})
	if err != nil {
		panic(err)
	}
	return state
}

func newTestCommand(
	t *testing.T,
	build func(config.Config) (Runtime, error),
) *urfavecli.Command {
	t.Helper()
	command, err := New(build)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return command
}

func writeLocalConfig(t *testing.T) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(configPath, []byte(`vault:
  driver: local
  localRoot: ./vault
index:
  driver: sqlite
  path: ./index.db
`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return configPath
}
