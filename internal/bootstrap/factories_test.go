package bootstrap

import (
	"testing"

	"git.casta.me/alberto/overmind/internal/config"
)

func TestNewLoggerRejectsUnsupportedLevel(t *testing.T) {
	if _, err := newLogger(config.Logging{Level: "verbose"}); err == nil {
		t.Fatal("NewLogger() error = nil")
	}
}

func TestNewBlobStoreRejectsUnsupportedDriver(t *testing.T) {
	if _, err := newBlobStore(config.Vault{Driver: "remote"}); err == nil {
		t.Fatal("NewBlobStore() error = nil")
	}
}
