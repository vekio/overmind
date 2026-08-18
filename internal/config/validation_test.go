package config

import (
	"strings"
	"testing"
)

func TestCLIConfigValidate(t *testing.T) {
	valid := validCLIConfig()
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := map[string]struct {
		modify func(*CLIConfig)
		want   string
	}{
		"logging": {
			modify: func(config *CLIConfig) { config.Logging.Level = "verbose" },
			want:   "logging.level",
		},
		"vault": {
			modify: func(config *CLIConfig) { config.Vault.RootPath = "" },
			want:   "vault.localRoot",
		},
		"index": {
			modify: func(config *CLIConfig) { config.Index.Driver = "json" },
			want:   "index.driver",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			config := validCLIConfig()
			test.modify(&config)
			if err := config.Validate(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestServerConfigValidate(t *testing.T) {
	valid := validServerConfig()
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := map[string]struct {
		modify func(*ServerConfig)
		want   string
	}{
		"logging": {
			modify: func(config *ServerConfig) { config.Logging.Level = "verbose" },
			want:   "logging.level",
		},
		"vault": {
			modify: func(config *ServerConfig) { config.Vault.Driver = "remote" },
			want:   "vault.driver",
		},
		"index": {
			modify: func(config *ServerConfig) { config.Index.Path = "" },
			want:   "index.path",
		},
		"HTTP address": {
			modify: func(config *ServerConfig) { config.HTTP.Address = "localhost" },
			want:   "http.address",
		},
		"HTTP port": {
			modify: func(config *ServerConfig) { config.HTTP.Address = "127.0.0.1:0" },
			want:   "http.address port",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			config := validServerConfig()
			test.modify(&config)
			if err := config.Validate(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func validCLIConfig() CLIConfig {
	return CLIConfig{
		Vault:   Vault{Driver: "local", RootPath: "./vault"},
		Logging: Logging{Level: "info"},
		Index:   Index{Driver: "sqlite", Path: "./index.db"},
	}
}

func validServerConfig() ServerConfig {
	return ServerConfig{
		Vault:   Vault{Driver: "local", RootPath: "./vault"},
		Logging: Logging{Level: "info"},
		HTTP:    HTTP{Address: "127.0.0.1:8080"},
		Index:   Index{Driver: "sqlite", Path: "./index.db"},
	}
}
