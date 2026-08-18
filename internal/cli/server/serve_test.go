package server

import (
	"context"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/config"
)

func TestLocalHTTPURL(t *testing.T) {
	tests := map[string]struct {
		address string
		want    string
	}{
		"loopback":       {address: "127.0.0.1:8080", want: "http://127.0.0.1:8080/"},
		"all interfaces": {address: "0.0.0.0:8080", want: "http://localhost:8080/"},
		"IPv6":           {address: "[::1]:8080", want: "http://[::1]:8080/"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := localHTTPURL(test.address); got != test.want {
				t.Fatalf("localHTTPURL(%q) = %q, want %q", test.address, got, test.want)
			}
		})
	}
}

func TestServeLoadsServerConfigurationAndValidatesOverride(t *testing.T) {
	built := false
	command := newTestCommand(t, func(config.ServerConfig) (Runtime, error) {
		built = true
		return runtimeStub{application: &app.Application{}}, nil
	})

	err := command.Run(context.Background(), []string{
		"overmind-server", "--config", writeServerConfig(t),
		"serve", "--address", "invalid",
	})
	if err == nil || !strings.Contains(err.Error(), "HTTP address") {
		t.Fatalf("Run() error = %v, want invalid HTTP address error", err)
	}
	if !built {
		t.Fatal("server application was not built")
	}
}
