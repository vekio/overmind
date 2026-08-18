package config

import (
	"fmt"
	"net"
	"strconv"
)

// Validate checks that the local CLI settings are usable.
func (config CLIConfig) Validate() error {
	if err := validateLogging(config.Logging); err != nil {
		return err
	}
	if err := validateLocalVault(config.Vault); err != nil {
		return err
	}
	return validateLocalIndex(config.Index)
}

// Validate checks that the server settings are usable.
func (config ServerConfig) Validate() error {
	if err := validateLogging(config.Logging); err != nil {
		return err
	}
	if err := validateLocalVault(config.Vault); err != nil {
		return err
	}
	if err := validateLocalIndex(config.Index); err != nil {
		return err
	}
	return validateHTTP(config.HTTP)
}

func validateLogging(logging Logging) error {
	switch logging.Level {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("unsupported logging.level %q", logging.Level)
	}
}

func validateHTTP(httpConfig HTTP) error {
	_, port, err := net.SplitHostPort(httpConfig.Address)
	if err != nil {
		return fmt.Errorf("invalid http.address %q: %w", httpConfig.Address, err)
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return fmt.Errorf("invalid http.address port %q", port)
	}
	return nil
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
