package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v4"
)

const defaultDirectoryName = "overmind"

func defaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, defaultDirectoryName, "config.yml"), nil
}

func defaultConfig() (Config, error) {
	dataDir, err := defaultDataDir()
	if err != nil {
		return Config{}, fmt.Errorf("resolve user data directory: %w", err)
	}
	dataPath := filepath.Join(dataDir, defaultDirectoryName)

	return Config{
		Vault: Vault{
			Driver:   "local",
			RootPath: filepath.Join(dataPath, "vault"),
		},
		Logging: Logging{
			Level: "info",
		},
		HTTP: HTTP{
			Address: "127.0.0.1:8080",
		},
		CLI: CLI{
			Mode: CLIModeLocal,
		},
		Index: Index{
			Driver: "sqlite",
			Path:   filepath.Join(dataPath, "index.db"),
		},
	}, nil
}

func defaultDataDir() (string, error) {
	if dataDir := os.Getenv("XDG_DATA_HOME"); dataDir != "" {
		if !filepath.IsAbs(dataDir) {
			return "", fmt.Errorf("XDG_DATA_HOME must be an absolute path")
		}
		return dataDir, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".local", "share"), nil
}

// WriteDefault writes a usable local configuration without replacing an
// existing file unless force is true.
func WriteDefault(path string, force bool) error {
	cfg, err := defaultConfig()
	if err != nil {
		return err
	}
	content, err := yaml.Dump(cfg, yaml.WithIndent(2))
	if err != nil {
		return fmt.Errorf("encode default config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	flags := os.O_WRONLY | os.O_CREATE
	if force {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	file, err := os.OpenFile(path, flags, 0o644)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("config file already exists %q: %w", path, os.ErrExist)
	}
	if err != nil {
		return err
	}
	defer file.Close()

	written, err := file.Write(content)
	if err != nil {
		return err
	}
	if written != len(content) {
		return io.ErrShortWrite
	}

	return nil
}
