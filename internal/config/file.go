package config

import (
	"fmt"

	configlib "github.com/vekio/config"
)

const (
	cliApplicationName    = "overmind"
	serverApplicationName = "overmind-server"
)

// NewCLIFile constructs the local CLI configuration file descriptor together
// with the complete defaults used when initializing it.
func NewCLIFile() (*configlib.ConfigFile[CLIConfig], CLIConfig, error) {
	defaults, err := defaultCLIConfig()
	if err != nil {
		return nil, CLIConfig{}, err
	}
	file, err := configlib.NewDefaultConfigFile[CLIConfig](cliApplicationName)
	if err != nil {
		return nil, CLIConfig{}, fmt.Errorf("create CLI configuration file: %w", err)
	}
	return file, defaults, nil
}

// NewServerFile constructs the server configuration file descriptor together
// with the complete defaults used when initializing it.
func NewServerFile() (*configlib.ConfigFile[ServerConfig], ServerConfig, error) {
	defaults, err := defaultServerConfig()
	if err != nil {
		return nil, ServerConfig{}, err
	}
	file, err := configlib.NewDefaultConfigFile[ServerConfig](serverApplicationName)
	if err != nil {
		return nil, ServerConfig{}, fmt.Errorf("create server configuration file: %w", err)
	}
	return file, defaults, nil
}
