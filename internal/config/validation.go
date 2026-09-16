package config

import (
	"fmt"
)

// Validate checks that the local application settings are usable.
func (config Config) Validate() error {
	if err := validateLocalVault(config.Vault); err != nil {
		return err
	}
	return validateLocalIndex(config.Index)
}

func validateLocalIndex(index Index) error {
	if index.Driver != "sqlite" {
		return fmt.Errorf("unsupported index.driver %q", index.Driver)
	}
	if index.Path == "" {
		return fmt.Errorf("index.path is required when index.driver is sqlite")
	}
	return nil
}

func validateLocalVault(vault Vault) error {
	if vault.Driver == "" {
		return fmt.Errorf("vault.driver is required in local CLI mode")
	}
	if vault.Driver != "local" {
		return fmt.Errorf("unsupported vault.driver %q", vault.Driver)
	}
	if vault.RootPath == "" {
		return fmt.Errorf("vault.localRoot is required when vault.driver is local")
	}
	return nil
}
