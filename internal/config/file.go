package config

import (
	"fmt"

	configlib "github.com/vekio/config"
)

const (
	applicationName = "overmind"
	fileName        = "config.yml"
)

// NewFile configures a YAML file handle with application defaults.
// It does not write the file; setup explicitly creates it after valid submission.
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
