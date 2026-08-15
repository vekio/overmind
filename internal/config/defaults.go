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

func defaultConfig() Config {
	return Config{
		Vault: Vault{
			Driver: "local",
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
			Path:   "./.overmind/index.db",
		},
	}
}

func WriteDefault(path string, force bool) error {
	content, err := yaml.Dump(defaultConfig(), yaml.WithIndent(2))
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
