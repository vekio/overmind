package config

import (
	"fmt"
	"strings"

	configlib "github.com/vekio/config"
)

// Settings contains the Overmind application settings.
type Settings struct {
	Mode      Mode   `yaml:"mode,omitempty"`
	VaultPath string `yaml:"vault,omitempty"`
}

// Mode identifies the storage backend selected by configuration.
type Mode string

// ModeLocal stores documents and the derived SQLite index in the local vault.
const ModeLocal Mode = "local"

func defaults() (Settings, error) {
	dataDir, err := configlib.DefaultDataDir(applicationName)
	if err != nil {
		return Settings{}, fmt.Errorf("resolve user data directory: %w", err)
	}
	return Settings{Mode: ModeLocal, VaultPath: dataDir}, nil
}

// Validate checks that the configuration is usable.
func (settings Settings) Validate() error {
	if settings.Mode != "" && settings.Mode != ModeLocal {
		return fmt.Errorf("unsupported mode %q: expected local", settings.Mode)
	}
	if strings.TrimSpace(settings.VaultPath) == "" {
		return fmt.Errorf("vault is required")
	}
	return nil
}
