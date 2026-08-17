package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/config"
)

func TestLocalHTTPURL(t *testing.T) {
	for name, test := range map[string]struct {
		address string
		want    string
	}{
		"loopback":       {address: "127.0.0.1:8080", want: "http://127.0.0.1:8080/"},
		"all interfaces": {address: "0.0.0.0:8080", want: "http://localhost:8080/"},
		"IPv6":           {address: "[::1]:8080", want: "http://[::1]:8080/"},
	} {
		t.Run(name, func(t *testing.T) {
			if got := localHTTPURL(test.address); got != test.want {
				t.Fatalf("localHTTPURL(%q) = %q, want %q", test.address, got, test.want)
			}
		})
	}
}

func TestServeLoadsApplicationAndValidatesAddressOverride(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(configPath, []byte("vault:\n  localRoot: ./vault\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	built := false
	command := NewCommand(func(config.Config) (Runtime, error) {
		built = true
		return runtimeStub{application: &app.Application{}}, nil
	})

	err := command.Run(context.Background(), []string{
		"overmind", "--config", configPath,
		"serve", "--address", "invalid",
	})
	if err == nil || !strings.Contains(err.Error(), "HTTP address") {
		t.Fatalf("Run() error = %v, want invalid HTTP address error", err)
	}
	if !built {
		t.Fatal("application was not built for serve command")
	}
}

func TestServeRejectsRemoteMode(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(configPath, []byte(`cli:
  mode: remote
  endpoint: https://overmind.example
`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	command := NewCommand(func(config.Config) (Runtime, error) {
		return runtimeStub{application: &app.Application{}}, nil
	})
	err := command.Run(context.Background(), []string{"overmind", "--config", configPath, "serve"})
	if err == nil || !strings.Contains(err.Error(), "requires cli.mode") {
		t.Fatalf("Run() error = %v, want local-mode error", err)
	}
}
