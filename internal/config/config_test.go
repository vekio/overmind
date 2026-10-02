package config_test

import (
	"strings"
	"testing"

	"github.com/vekio/overmind/internal/config"
)

func TestSettingsValidateLocalModeAndVault(t *testing.T) {
	for _, mode := range []config.Mode{"", config.ModeLocal} {
		if err := (config.Settings{Mode: mode, VaultPath: "/tmp/vault"}).Validate(); err != nil {
			t.Fatalf("valid settings with mode %q: %v", mode, err)
		}
	}
	for _, tc := range []struct {
		settings config.Settings
		want     string
	}{
		{config.Settings{Mode: "api", VaultPath: "/tmp/vault"}, "unsupported mode"},
		{config.Settings{Mode: config.ModeLocal, VaultPath: "  "}, "vault is required"},
	} {
		if err := tc.settings.Validate(); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("Validate(%+v) = %v, want %q", tc.settings, err, tc.want)
		}
	}
}
