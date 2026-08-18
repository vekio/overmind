package local

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/config"
	"git.casta.me/alberto/overmind/internal/domain"
)

type createPageHandlerStub struct {
	command createpage.CreatePageCommand
	err     error
}

func (handler *createPageHandlerStub) Handle(_ context.Context, command createpage.CreatePageCommand) (createpage.CreatePageResult, error) {
	handler.command = command
	if handler.err != nil {
		return createpage.CreatePageResult{}, handler.err
	}
	id, err := domain.NewDocumentID("page-id")
	return createpage.CreatePageResult{ID: id, Path: "page/knowledge/go/first-page.adoc"}, err
}

func TestCreatePagePreservesUseCaseError(t *testing.T) {
	useCaseErr := errors.New("store page: blob already exists")
	handler := &createPageHandlerStub{err: useCaseErr}
	state := newTestApplicationState(runtimeStub{application: &app.Application{
		Commands: app.Commands{CreatePage: handler},
	}})

	err := newPageCommand(state).Run(context.Background(), []string{"page", "Page"})
	if !errors.Is(err, useCaseErr) || !strings.Contains(err.Error(), "blob already exists") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestCreatePageLoadsConfigAndExecutesUseCase(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(configPath, []byte(`vault:
  driver: local
  localRoot: ./vault
logging:
  level: info
index:
  driver: sqlite
  path: ./index.db
`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	handler := &createPageHandlerStub{}
	var selectedConfig config.CLIConfig
	command := newTestCommand(t, func(cfg config.CLIConfig) (Runtime, error) {
		selectedConfig = cfg
		return runtimeStub{application: &app.Application{
			Commands: app.Commands{CreatePage: handler},
		}}, nil
	})
	var output bytes.Buffer
	command.Writer = &output

	err := command.Run(context.Background(), []string{
		"overmind", "--config", configPath, "--debug",
		"page", "First page", "--area", "Knowledge/Go", "--tag", "Go", "--tag", "DDD",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if selectedConfig.Vault.RootPath != "./vault" {
		t.Fatalf("selected root = %q", selectedConfig.Vault.RootPath)
	}
	if selectedConfig.Logging.Level != "debug" {
		t.Fatalf("selected logging level = %q", selectedConfig.Logging.Level)
	}
	if handler.command.Title != "First page" || handler.command.Area != "Knowledge/Go" || len(handler.command.Tags) != 2 || handler.command.Tags[0] != "Go" || handler.command.Tags[1] != "DDD" {
		t.Fatalf("use-case command = %+v", handler.command)
	}
	if output.String() != "page/knowledge/go/first-page.adoc\n" {
		t.Fatalf("output = %q", output.String())
	}
}

func TestCreatePageCreatesAndLoadsDefaultConfig(t *testing.T) {
	configHome := filepath.Join(t.TempDir(), "config")
	dataHome := filepath.Join(t.TempDir(), "data")
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_DATA_HOME", dataHome)

	handler := &createPageHandlerStub{}
	var selectedConfig config.CLIConfig
	command := newTestCommand(t, func(cfg config.CLIConfig) (Runtime, error) {
		selectedConfig = cfg
		return runtimeStub{application: &app.Application{
			Commands: app.Commands{CreatePage: handler},
		}}, nil
	})
	command.Writer = &bytes.Buffer{}

	if err := command.Run(context.Background(), []string{"overmind", "page", "First page"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if want := filepath.Join(dataHome, "overmind", "vault"); selectedConfig.Vault.RootPath != want {
		t.Fatalf("Vault.RootPath = %q, want %q", selectedConfig.Vault.RootPath, want)
	}
	if _, err := os.Stat(filepath.Join(configHome, "overmind", "config.yml")); err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
}

func TestCreatePageDoesNotCreateExplicitMissingConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "missing.yml")
	built := false
	command := newTestCommand(t, func(config.CLIConfig) (Runtime, error) {
		built = true
		return runtimeStub{application: &app.Application{}}, nil
	})

	err := command.Run(context.Background(), []string{
		"overmind", "--config", configPath, "page", "First page",
	})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Run() error = %v, want %v", err, os.ErrNotExist)
	}
	if built {
		t.Fatal("application was built with missing explicit configuration")
	}
	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat() error = %v, want %v", err, os.ErrNotExist)
	}
}

func TestCreatePageRequiresTitleArgument(t *testing.T) {
	err := newPageCommand(&applicationState{}).Run(context.Background(), []string{"page"})
	if err == nil {
		t.Fatalf("Run() error = %v, want missing title error", err)
	}
}

func TestCreatePageRejectsArgumentsAfterTitle(t *testing.T) {
	err := newPageCommand(&applicationState{}).Run(context.Background(), []string{"page", "First", "page"})
	if err == nil || !strings.Contains(err.Error(), "unexpected arguments") {
		t.Fatalf("Run() error = %v, want unexpected arguments error", err)
	}
}
