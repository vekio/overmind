package config

import (
	"fmt"

	configlib "github.com/vekio/config"
)

const (
	applicationName = "overmind"
	fileName        = "config.yml"
)

// NewFile creates the application configuration file with its defaults.
func NewFile() (*configlib.ConfigFile[Settings], error) {
	defaultSettings, err := defaults()
	if err != nil {
		return nil, err
	}

	file, err := configlib.NewYAMLConfigFile(
		applicationName,
		fileName,
		configlib.Default(defaultSettings),
	)
	if err != nil {
		return nil, fmt.Errorf("create configuration file: %w", err)
	}
	return file, nil
}
