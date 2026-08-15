package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
)

const (
	CLIModeLocal  = "local"
	CLIModeRemote = "remote"
)

type Config struct {
	Vault   Vault   `yaml:"vault"`
	Logging Logging `yaml:"logging"`
	HTTP    HTTP    `yaml:"http"`
	CLI     CLI     `yaml:"cli"`
	Index   Index   `yaml:"index"`
}

type Vault struct {
	Driver   string `yaml:"driver"`
	RootPath string `yaml:"localRoot"`
}

type Logging struct {
	Level string `yaml:"level"`
}

type HTTP struct {
	Address string `yaml:"address"`
}

type CLI struct {
	Mode     string `yaml:"mode"`
	Endpoint string `yaml:"endpoint,omitempty"`
}

type Index struct {
	Driver string `yaml:"driver"`
	Path   string `yaml:"path"`
}

func (c Config) Validate() error {
	switch c.CLI.Mode {
	case CLIModeLocal:
		if err := c.validateLocalVault(); err != nil {
			return err
		}
		if err := c.validateLocalIndex(); err != nil {
			return err
		}
	case CLIModeRemote:
		if err := validateEndpoint(c.CLI.Endpoint); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported cli.mode %q", c.CLI.Mode)
	}

	_, port, err := net.SplitHostPort(c.HTTP.Address)
	if err != nil {
		return fmt.Errorf("invalid http.address %q: %w", c.HTTP.Address, err)
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return fmt.Errorf("invalid http.address port %q", port)
	}

	return nil
}

func (c Config) validateLocalIndex() error {
	if c.Index.Driver != "sqlite" {
		return fmt.Errorf("unsupported index.driver %q", c.Index.Driver)
	}
	if c.Index.Path == "" {
		return fmt.Errorf("index.path is required when index.driver is sqlite")
	}
	return nil
}

func (c Config) validateLocalVault() error {
	if c.Vault.Driver == "" {
		return fmt.Errorf("vault.driver is required in local CLI mode")
	}
	if c.Vault.Driver != "local" {
		return fmt.Errorf("unsupported vault.driver %q", c.Vault.Driver)
	}
	if c.Vault.RootPath == "" {
		return fmt.Errorf("vault.localRoot is required when vault.driver is local")
	}
	return nil
}

func validateEndpoint(endpoint string) error {
	parsed, err := url.Parse(endpoint)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("invalid cli.endpoint %q", endpoint)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("cli.endpoint must not contain query or fragment")
	}
	return nil
}
