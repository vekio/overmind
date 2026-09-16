package bootstrap

import (
	"testing"

	"git.casta.me/alberto/overmind/internal/config"
)

func TestNewBlobStoreRejectsUnsupportedDriver(t *testing.T) {
	if _, err := newBlobStore(config.Vault{Driver: "remote"}); err == nil {
		t.Fatal("NewBlobStore() error = nil")
	}
}
