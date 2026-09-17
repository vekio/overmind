package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	urfavecli "github.com/urfave/cli/v3"
)

func newTestCommand(t *testing.T) *urfavecli.Command {
	t.Helper()
	command, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return command
}

func writeLocalConfig(t *testing.T) (string, string) {
	t.Helper()
	dataDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.yml")
	content := fmt.Appendf(nil, "dataDir: %s\n", dataDir)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return configPath, dataDir
}
