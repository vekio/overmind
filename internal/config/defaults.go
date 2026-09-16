package config

import (
	"fmt"
	"path/filepath"

	configlib "github.com/vekio/config"
)

const dataApplicationName = "overmind"

// Default returns the default local application configuration.
func Default() (Config, error) {
	vault, index, err := defaultLocalStorage()
	if err != nil {
		return Config{}, err
	}
	return Config{
		Vault: vault,
		Index: index,
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
