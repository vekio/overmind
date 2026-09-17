package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupCreatesDefaultConfigurationWithoutOpeningApplication(t *testing.T) {
	configHome := filepath.Join(t.TempDir(), "config")
	dataHome := filepath.Join(t.TempDir(), "data")
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_DATA_HOME", dataHome)

	command := newTestCommand(t)
	var output bytes.Buffer
	command.Writer = &output
	if err := command.Run(context.Background(), []string{"overmind", "setup"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	configPath := filepath.Join(configHome, "overmind", "config.yml")
	if output.String() != "Configuration: "+configPath+"\n" {
		t.Fatalf("output = %q", output.String())
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Stat(config) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dataHome, "overmind")); !os.IsNotExist(err) {
		t.Fatalf("data directory was created during setup: %v", err)
	}
}
