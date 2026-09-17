package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreatePageLoadsConfigAndCreatesDocument(t *testing.T) {
	configPath, dataDir := writeLocalConfig(t)
	command := newTestCommand(t)
	var output bytes.Buffer
	command.Writer = &output

	err := command.Run(context.Background(), []string{
		"overmind", "--config", configPath,
		"page", "First page", "--area", "Knowledge/Go", "--tag", "Go", "--tag", "DDD",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	id := strings.TrimSpace(output.String())
	content, err := os.ReadFile(filepath.Join(dataDir, "documents", id+".adoc"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(content)
	for _, want := range []string{"= First page", ":overmind-area: knowledge/go", ":overmind-tags: go, ddd"} {
		if !strings.Contains(got, want) {
			t.Fatalf("document does not contain %q:\n%s", want, got)
		}
	}
}

func TestCreatePageLoadsConfigurationCreatedBySetup(t *testing.T) {
	configHome := filepath.Join(t.TempDir(), "config")
	dataHome := filepath.Join(t.TempDir(), "data")
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_DATA_HOME", dataHome)

	setup := newTestCommand(t)
	setup.Writer = &bytes.Buffer{}
	if err := setup.Run(context.Background(), []string{"overmind", "setup"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	command := newTestCommand(t)
	var output bytes.Buffer
	command.Writer = &output

	if err := command.Run(context.Background(), []string{"overmind", "page", "First page"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(configHome, "overmind", "config.yml")); err != nil {
		t.Fatalf("Stat(config) error = %v", err)
	}
	id := strings.TrimSpace(output.String())
	if _, err := os.Stat(filepath.Join(dataHome, "overmind", "documents", id+".adoc")); err != nil {
		t.Fatalf("Stat(document) error = %v", err)
	}
}

func TestCreatePageDoesNotCreateExplicitMissingConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "missing.yml")
	command := newTestCommand(t)

	err := command.Run(context.Background(), []string{
		"overmind", "--config", configPath, "page", "First page",
	})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Run() error = %v, want %v", err, os.ErrNotExist)
	}
	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat() error = %v, want %v", err, os.ErrNotExist)
	}
}

func TestCreatePageRequiresSetupWhenDefaultConfigurationIsMissing(t *testing.T) {
	configHome := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", configHome)

	command := newTestCommand(t)
	err := command.Run(context.Background(), []string{"overmind", "page", "First page"})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Run() error = %v, want %v", err, os.ErrNotExist)
	}
	if _, err := os.Stat(filepath.Join(configHome, "overmind", "config.yml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("configuration was created without setup: %v", err)
	}
}

func TestCreatePageRequiresTitleArgument(t *testing.T) {
	command := newTestCommand(t)
	err := command.Run(context.Background(), []string{"overmind", "page"})
	if err == nil {
		t.Fatalf("Run() error = %v, want missing title error", err)
	}
}

func TestCreatePageRejectsArgumentsAfterTitle(t *testing.T) {
	command := newTestCommand(t)
	err := command.Run(context.Background(), []string{"overmind", "page", "First", "page"})
	if err == nil || !strings.Contains(err.Error(), "unexpected arguments") {
		t.Fatalf("Run() error = %v, want unexpected arguments error", err)
	}
}
