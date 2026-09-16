package config

import (
	"fmt"

	configlib "github.com/vekio/config"
)

const (
	applicationName = "overmind"
	fileName        = "config.yml"
)

// NewFile constructs the local configuration file descriptor together with
// the complete defaults used when initializing it.
func NewFile() (*configlib.ConfigFile[Config], Config, error) {
	defaults, err := Default()
	if err != nil {
		return nil, Config{}, err
	}
	file, err := configlib.NewYAMLConfigFile[Config](applicationName, fileName)
	if err != nil {
		return nil, Config{}, fmt.Errorf("create configuration file: %w", err)
	}
	return file, defaults, nil
}
