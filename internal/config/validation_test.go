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

	invalid := Config{}
	if err := invalid.Validate(); err == nil || !strings.Contains(err.Error(), "dataDir") {
		t.Fatalf("Validate() error = %v, want dataDir", err)
	}
}

func validConfig() Config {
	return Config{DataDir: "./data"}
}
