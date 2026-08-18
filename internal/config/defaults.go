package config

import (
	"fmt"
	"path/filepath"

	configlib "github.com/vekio/config"
)

const dataApplicationName = "overmind"

func defaultCLIConfig() (CLIConfig, error) {
	vault, index, err := defaultLocalStorage()
	if err != nil {
		return CLIConfig{}, err
	}
	return CLIConfig{
		Vault:   vault,
		Logging: Logging{Level: "info"},
		Index:   index,
	}, nil
}

func defaultServerConfig() (ServerConfig, error) {
	vault, index, err := defaultLocalStorage()
	if err != nil {
		return ServerConfig{}, err
	}
	return ServerConfig{
		Vault:   vault,
		Logging: Logging{Level: "info"},
		HTTP:    HTTP{Address: "127.0.0.1:8080"},
		Index:   index,
	}, nil
}

func defaultLocalStorage() (Vault, Index, error) {
	dataPath, err := configlib.DefaultDataDir(dataApplicationName)
	if err != nil {
		return Vault{}, Index{}, fmt.Errorf("resolve user data directory: %w", err)
	}
	return Vault{
			Driver:   "local",
			RootPath: filepath.Join(dataPath, "vault"),
		}, Index{
			Driver: "sqlite",
			Path:   filepath.Join(dataPath, "index.db"),
		}, nil
}
