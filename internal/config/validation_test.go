package config

import (
	"strings"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	valid := validConfig()
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := map[string]struct {
		modify func(*Config)
		want   string
	}{
		"vault": {
			modify: func(config *Config) { config.Vault.RootPath = "" },
			want:   "vault.localRoot",
		},
		"index": {
			modify: func(config *Config) { config.Index.Driver = "json" },
			want:   "index.driver",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			config := validConfig()
			test.modify(&config)
			if err := config.Validate(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func validConfig() Config {
	return Config{
		Vault: Vault{Driver: "local", RootPath: "./vault"},
		Index: Index{Driver: "sqlite", Path: "./index.db"},
	}
}
