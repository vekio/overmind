package cli

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/app/rebuildindex"
	"git.casta.me/alberto/overmind/internal/config"
	"git.casta.me/alberto/overmind/internal/domain"
)

type createPageHandlerStub struct {
	command createpage.CreatePageCommand
}

type getDocumentHandlerStub struct {
	query  getdocument.GetDocumentQuery
	result getdocument.GetDocumentResult
}

func (handler *getDocumentHandlerStub) Handle(_ context.Context, query getdocument.GetDocumentQuery) (getdocument.GetDocumentResult, error) {
	handler.query = query
	return handler.result, nil
}

type rebuildIndexHandlerStub struct {
	called bool
	result rebuildindex.RebuildIndexResult
}

func (handler *rebuildIndexHandlerStub) Handle(context.Context, rebuildindex.RebuildIndexCommand) (rebuildindex.RebuildIndexResult, error) {
	handler.called = true
	return handler.result, nil
}

type runtimeStub struct {
	application *app.Application
}

func (runtime runtimeStub) Application() *app.Application { return runtime.application }

func (runtime runtimeStub) Logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (runtime runtimeStub) Close() error { return nil }

func (handler *createPageHandlerStub) Handle(_ context.Context, command createpage.CreatePageCommand) (createpage.CreatePageResult, error) {
	handler.command = command
	id, err := domain.NewDocumentID("page-id")
	return createpage.CreatePageResult{ID: id}, err
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
		"overmind", "--config", configPath,
		"page", "First page", "--area", "Knowledge/Go",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if selectedConfig.Vault.RootPath != "./vault" {
		t.Fatalf("selected root = %q", selectedConfig.Vault.RootPath)
	}
	if handler.command.Title != "First page" || handler.command.Area != "Knowledge/Go" {
		t.Fatalf("use-case command = %+v", handler.command)
	}
	if output.String() != "page-id\n" {
		t.Fatalf("output = %q", output.String())
	}
}

func TestCreatePageRequiresTitleArgument(t *testing.T) {
	err := newPageCommand(&applicationState{}).Run(context.Background(), []string{"page"})
	if err == nil || !strings.Contains(err.Error(), "title") {
		t.Fatalf("Run() error = %v, want missing title error", err)
	}
}

func TestDocumentGetsDocumentByID(t *testing.T) {
	configPath := writeLocalConfig(t)
	id, _ := domain.NewDocumentID("page-id")
	handler := &getDocumentHandlerStub{result: getdocument.GetDocumentResult{
		ID: id, Content: []byte("= Page\n"),
	}}
	command := NewCommand(func(config.Config) (Runtime, error) {
		return runtimeStub{application: &app.Application{
			Queries: app.Queries{GetDocument: handler},
		}}, nil
	})
	var output bytes.Buffer
	command.Writer = &output

	if err := command.Run(context.Background(), []string{"overmind", "--config", configPath, "document", "page-id"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if handler.query.ID != "page-id" || output.String() != "= Page\n" {
		t.Fatalf("query = %+v, output = %q", handler.query, output.String())
	}
}

func TestIndexRebuildExecutesLocalUseCase(t *testing.T) {
	configPath := writeLocalConfig(t)
	handler := &rebuildIndexHandlerStub{result: rebuildindex.RebuildIndexResult{Documents: 2}}
	command := NewCommand(func(config.Config) (Runtime, error) {
		return runtimeStub{application: &app.Application{
			Commands: app.Commands{RebuildIndex: handler},
		}}, nil
	})
	var output bytes.Buffer
	command.Writer = &output

	if err := command.Run(context.Background(), []string{"overmind", "--config", configPath, "index", "rebuild"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !handler.called || output.String() != "Indexed 2 documents\n" {
		t.Fatalf("called = %t, output = %q", handler.called, output.String())
	}
}

func writeLocalConfig(t *testing.T) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(configPath, []byte("vault:\n  localRoot: ./vault\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return configPath
}

func TestAsciiDocCommandDoesNotLoadApplicationConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.adoc")
	if err := os.WriteFile(path, []byte("plain text\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	built := false
	command := NewCommand(func(config.Config) (Runtime, error) {
		built = true
		return runtimeStub{application: &app.Application{
			Commands: app.Commands{CreatePage: &createPageHandlerStub{}},
		}}, nil
	})
	command.Writer = &bytes.Buffer{}
	if err := command.Run(context.Background(), []string{"overmind", "asciidoc", "lexer", path}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if built {
		t.Fatal("application was built for standalone asciidoc command")
	}
}

func TestServeLoadsApplicationAndValidatesAddressOverride(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(configPath, []byte(`vault:
  localRoot: ./vault
`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	built := false
	command := NewCommand(func(config.Config) (Runtime, error) {
		built = true
		return runtimeStub{application: &app.Application{
			Commands: app.Commands{CreatePage: &createPageHandlerStub{}},
		}}, nil
	})

	err := command.Run(context.Background(), []string{
		"overmind", "--config", configPath,
		"serve", "--address", "invalid",
	})
	if err == nil || !strings.Contains(err.Error(), "HTTP address") {
		t.Fatalf("Run() error = %v, want invalid HTTP address error", err)
	}
	if !built {
		t.Fatal("application was not built for serve command")
	}
}

func TestServeRejectsRemoteMode(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(configPath, []byte(`cli:
  mode: remote
  endpoint: https://overmind.example
`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	command := NewCommand(func(config.Config) (Runtime, error) {
		return runtimeStub{application: &app.Application{
			Commands: app.Commands{CreatePage: &createPageHandlerStub{}},
		}}, nil
	})
	err := command.Run(context.Background(), []string{"overmind", "--config", configPath, "serve"})
	if err == nil || !strings.Contains(err.Error(), "requires cli.mode") {
		t.Fatalf("Run() error = %v, want local-mode error", err)
	}
}
