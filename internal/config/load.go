package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"go.yaml.in/yaml/v4"
)

func Load() (Config, error) {
	path, err := defaultPath()
	if err != nil {
		return Config{}, err
	}

	cfg, err := LoadFrom(path)
	if !errors.Is(err, os.ErrNotExist) {
		return cfg, err
	}

	if err := WriteDefault(path, false); err != nil && !errors.Is(err, os.ErrExist) {
		return Config{}, fmt.Errorf("create default config file %q: %w", path, err)
	}
	return LoadFrom(path)
}

func LoadFrom(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("config file not found %q: %w", path, os.ErrNotExist)
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", path, err)
	}

	cfg, err := defaultConfig()
	if err != nil {
		return Config{}, err
	}

	loader, err := yaml.NewLoader(
		bytes.NewReader(content),
		yaml.WithKnownFields(),
		yaml.WithUniqueKeys(),
	)
	if err != nil {
		return Config{}, fmt.Errorf("create YAML loader: %w", err)
	}
	if err := loader.Load(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file %q: %w", path, err)
	}

	var extraDocument any
	if err := loader.Load(&extraDocument); err == nil {
		return Config{}, fmt.Errorf("parse config file %q: multiple YAML documents are not allowed", path)
	} else if !errors.Is(err, io.EOF) {
		return Config{}, fmt.Errorf("parse config file %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
