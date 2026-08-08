package config

import "fmt"

type Config struct {
	Vault   Vault   `yaml:"vault"`
	Logging Logging `yaml:"logging"`
}

type Vault struct {
	Driver   string `yaml:"driver"`
	RootPath string `yaml:"localRoot"`
}

type Logging struct {
	Level string `yaml:"level"`
}

func (c Config) Validate() error {
	if c.Vault.Driver == "" {
		return fmt.Errorf("vault.driver is required")
	}

	if c.Vault.Driver == "local" && c.Vault.RootPath == "" {
		return fmt.Errorf("vault.localRoot is required when vault.driver is local")
	}

	return nil
}
