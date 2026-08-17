package cli

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
	state := &applicationState{runtime: runtimeStub{application: &app.Application{
		Commands: app.Commands{CreatePage: handler},
	}}}

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
`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	handler := &createPageHandlerStub{}
	var selectedConfig config.Config
	command := NewCommand(func(cfg config.Config) (Runtime, error) {
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
