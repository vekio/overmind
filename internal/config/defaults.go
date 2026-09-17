package config

import (
	"fmt"

	configlib "github.com/vekio/config"
)

const dataApplicationName = "overmind"

// Default returns the default local application configuration.
func Default() (Config, error) {
	dataDir, err := configlib.DefaultDataDir(dataApplicationName)
	if err != nil {
		return Config{}, fmt.Errorf("resolve user data directory: %w", err)
	}
	return Config{DataDir: dataDir}, nil
}
