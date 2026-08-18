package server

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/config"
)

func TestConfigInitDoesNotBuildServerApplication(t *testing.T) {
	configHome := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "data"))

	built := false
	command := newTestCommand(t, func(config.ServerConfig) (Runtime, error) {
		built = true
		return runtimeStub{application: &app.Application{}}, nil
	})
	var output bytes.Buffer
	command.Writer = &output

	if err := command.Run(context.Background(), []string{"overmind-server", "config", "init"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	wantPath := filepath.Join(configHome, "overmind-server", "config.yml")
	if output.String() != wantPath+"\n" {
		t.Fatalf("output = %q, want %q", output.String(), wantPath+"\n")
	}
	if built {
		t.Fatal("server application was built for config command")
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
}
