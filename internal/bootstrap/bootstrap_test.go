package bootstrap

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/config"
)

func TestContainerCreatesAnAsciiDocPage(t *testing.T) {
	root := t.TempDir()
	container, err := NewContainer(config.Config{
		Vault:   config.Vault{Driver: "local", RootPath: root},
		Logging: config.Logging{Level: "info"},
		HTTP:    config.HTTP{Address: "127.0.0.1:8080"},
		CLI:     config.CLI{Mode: config.CLIModeLocal},
		Index:   config.Index{Driver: "sqlite", Path: filepath.Join(root, "index.db")},
	})
	if err != nil {
		t.Fatalf("NewContainer() error = %v", err)
	}
	t.Cleanup(func() {
		if err := container.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	if container.log == nil {
		t.Fatal("Container.Log = nil")
	}
	if container.app == nil {
		t.Fatal("Container.App = nil")
	}

	result, err := container.app.Commands.CreatePage.Handle(context.Background(), createpage.CreatePageCommand{
		Title: "First page",
		Area:  "Knowledge",
	})
	if err != nil {
		t.Fatalf("CreatePage.Handle() error = %v", err)
	}
	if result.ID.String() == "" {
		t.Fatal("CreatePage.Handle() returned an empty ID")
	}
	content, err := os.ReadFile(filepath.Join(root, "page", "knowledge", "first-page.adoc"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := string(content); !strings.HasPrefix(got, "= First page\n:overmind-id: ") {
		t.Fatalf("page content = %q", got)
	}
	indexed, err := container.app.Queries.GetDocument.Handle(context.Background(), getdocument.GetDocumentQuery{ID: result.ID.String()})
	if err != nil {
		t.Fatalf("GetDocument.Handle() error = %v", err)
	}
	if indexed.Path != "page/knowledge/first-page.adoc" || indexed.Kind != "page" || indexed.Title != "First page" || indexed.Attributes["area"] != "knowledge" || string(indexed.Content) != string(content) {
		t.Fatalf("indexed document = %+v", indexed)
	}
}

func TestContainerBuildsRemoteApplicationWithoutLocalVault(t *testing.T) {
	container, err := NewContainer(config.Config{
		Logging: config.Logging{Level: "info"},
		HTTP:    config.HTTP{Address: "127.0.0.1:8080"},
		CLI: config.CLI{
			Mode:     config.CLIModeRemote,
			Endpoint: "https://overmind.example",
		},
	})
	if err != nil {
		t.Fatalf("NewContainer() error = %v", err)
	}
	if container.app == nil || container.app.Commands.CreatePage == nil || container.app.Queries.GetDocument == nil {
		t.Fatal("remote application is incomplete")
	}
}
