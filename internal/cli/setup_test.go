package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/vekio/overmind/internal/config"
)

func TestNormalizeVaultPathRejectsFilesAndNormalizesInput(t *testing.T) {
	root := t.TempDir()
	path, err := appconfig.NormalizeVaultPath("  " + root + "/  ")
	if err != nil || path != root {
		t.Fatalf("directory path = %q, %v", path, err)
	}
	filePath := filepath.Join(root, "file")
	if err := os.WriteFile(filePath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{" ", filePath, "~someone/vault"} {
		if _, err := appconfig.NormalizeVaultPath(value); err == nil {
			t.Fatalf("invalid vault path %q accepted", value)
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	path, err = appconfig.NormalizeVaultPath("~/overmind-test-vault")
	if err != nil || path != filepath.Join(home, "overmind-test-vault") {
		t.Fatalf("expanded home path = %q, %v", path, err)
	}
}

func TestSetupCreatesConfigOnceWithNormalizedVault(t *testing.T) {
	root := t.TempDir()
	configFile, err := appconfig.NewFile()
	if err != nil {
		t.Fatal(err)
	}
	if err := configFile.SetPath(filepath.Join(root, "config.yml")); err != nil {
		t.Fatal(err)
	}
	called := 0
	prompt := func(_ context.Context, defaults appconfig.Settings) (appconfig.Settings, error) {
		if defaults != configFile.Defaults() {
			t.Fatalf("prompt defaults=%+v, want=%+v", defaults, configFile.Defaults())
		}
		called++
		return appconfig.Settings{Mode: appconfig.ModeLocal, VaultPath: "  " + filepath.Join(root, "vault") + "/  "}, nil
	}
	var output bytes.Buffer
	command := newSetupCommand(configFile, prompt)
	command.Writer = &output
	if err := command.Run(context.Background(), []string{"setup"}); err != nil {
		t.Fatal(err)
	}
	settings, err := configFile.Load()
	if err != nil || settings.Mode != appconfig.ModeLocal || settings.VaultPath != filepath.Join(root, "vault") {
		t.Fatalf("saved settings = %+v, %v", settings, err)
	}
	if output.String() != configFile.Path()+"\n" {
		t.Fatalf("setup output = %q", output.String())
	}
	command = newSetupCommand(configFile, prompt)
	command.Writer = io.Discard
	if err := command.Run(context.Background(), []string{"setup"}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("second setup = %v", err)
	}
	if called != 1 {
		t.Fatalf("prompt called %d times", called)
	}
}

func TestSetupDoesNotCreateConfigOnCancellationOrInvalidInput(t *testing.T) {
	cancelled := errors.New("cancelled prompt")
	for _, tc := range []struct {
		name     string
		settings appconfig.Settings
		err      error
		args     []string
	}{
		{name: "cancel", err: cancelled},
		{name: "mode", settings: appconfig.Settings{Mode: "api", VaultPath: t.TempDir()}},
		{name: "vault", settings: appconfig.Settings{Mode: appconfig.ModeLocal}},
		{name: "arguments", args: []string{"setup", "extra"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			configFile, err := appconfig.NewFile()
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "config.yml")
			if err := configFile.SetPath(path); err != nil {
				t.Fatal(err)
			}
			calls := 0
			command := newSetupCommand(configFile, func(context.Context, appconfig.Settings) (appconfig.Settings, error) {
				calls++
				return tc.settings, tc.err
			})
			command.Writer = io.Discard
			args := tc.args
			if len(args) == 0 {
				args = []string{"setup"}
			}
			err = command.Run(context.Background(), args)
			if err == nil {
				t.Fatal("invalid or cancelled setup succeeded")
			}
			if tc.err != nil && !errors.Is(err, tc.err) {
				t.Fatal("prompt error was not preserved")
			}
			if tc.name == "arguments" && calls != 0 {
				t.Fatal("unexpected arguments opened prompt")
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("setup created a file: %v", err)
			}
		})
	}
}

func TestSetupUsesRootConfigFlagWithoutLoadingApplication(t *testing.T) {
	configFile, err := appconfig.NewFile()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	configPath := filepath.Join(root, "custom", "config.yml")
	defaults := configFile.Defaults()
	command := New(configFile, func(context.Context) (Client, error) {
		t.Fatal("setup loaded application before configuration existed")
		return nil, nil
	})
	command.Command("setup").Action = newSetupCommand(configFile, func(_ context.Context, got appconfig.Settings) (appconfig.Settings, error) {
		if got != defaults || configFile.Path() != configPath {
			t.Fatal("setup lost defaults or custom config path")
		}
		got.VaultPath = filepath.Join(root, "vault")
		return got, nil
	}).Action
	command.Writer, command.ErrWriter = io.Discard, io.Discard
	if err := command.Run(context.Background(), []string{"overmind", "--config", configPath, "setup"}); err != nil {
		t.Fatal(err)
	}
	loaded, err := configFile.Load()
	if err != nil || loaded.Mode != defaults.Mode || loaded.VaultPath != filepath.Join(root, "vault") {
		t.Fatalf("created configuration=%+v, error=%v", loaded, err)
	}
}
