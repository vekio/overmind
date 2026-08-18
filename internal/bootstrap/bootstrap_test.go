package bootstrap

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/app/listdocuments"
	"git.casta.me/alberto/overmind/internal/app/updatedocument"
	"git.casta.me/alberto/overmind/internal/config"
)

func TestContainerCreatesAnAsciiDocPage(t *testing.T) {
	root := t.TempDir()
	container, err := NewCLIContainer(config.CLIConfig{
		Vault:   config.Vault{Driver: "local", RootPath: root},
		Logging: config.Logging{Level: "info"},
		Index:   config.Index{Driver: "sqlite", Path: filepath.Join(root, "index.db")},
	})
	if err != nil {
		t.Fatalf("NewCLIContainer() error = %v", err)
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
	retrieved, err := container.app.Queries.GetDocument.Handle(context.Background(), getdocument.GetDocumentQuery{
		Path: "page/knowledge/first-page.adoc",
	})
	if err != nil {
		t.Fatalf("GetDocument.Handle() error = %v", err)
	}
	if retrieved.Path != "page/knowledge/first-page.adoc" || string(retrieved.Content) != string(content) || retrieved.Revision == "" {
		t.Fatalf("retrieved document = %+v", retrieved)
	}
	editedContent := []byte(strings.Replace(string(retrieved.Content), "= First page", "= Updated page", 1))
	updated, err := container.app.Commands.UpdateDocument.Handle(context.Background(), updatedocument.UpdateDocumentCommand{
		Path:             retrieved.Path,
		Content:          editedContent,
		ExpectedRevision: retrieved.Revision,
	})
	if err != nil {
		t.Fatalf("UpdateDocument.Handle() error = %v", err)
	}
	if updated.Path != retrieved.Path || updated.Revision == "" || updated.Revision == retrieved.Revision {
		t.Fatalf("updated document = %+v", updated)
	}
	storedContent, err := os.ReadFile(filepath.Join(root, "page", "knowledge", "first-page.adoc"))
	if err != nil || string(storedContent) != string(editedContent) {
		t.Fatalf("stored edited content = %q, error = %v", storedContent, err)
	}
	listed, err := container.app.Queries.ListDocuments.Handle(context.Background(), listdocuments.ListDocumentsQuery{
		Type: "page",
	})
	if err != nil {
		t.Fatalf("ListDocuments.Handle() error = %v", err)
	}
	if len(listed.Documents) != 1 || listed.Documents[0].Path != "page/knowledge/first-page.adoc" {
		t.Fatalf("listed documents = %+v", listed.Documents)
	}
}

func TestServerContainerBuildsApplication(t *testing.T) {
	root := t.TempDir()
	container, err := NewServerContainer(config.ServerConfig{
		Vault:   config.Vault{Driver: "local", RootPath: root},
		Logging: config.Logging{Level: "info"},
		HTTP:    config.HTTP{Address: "127.0.0.1:8080"},
		Index:   config.Index{Driver: "sqlite", Path: filepath.Join(root, "index.db")},
	})
	if err != nil {
		t.Fatalf("NewServerContainer() error = %v", err)
	}
	t.Cleanup(func() { _ = container.Close() })
	if container.Application() == nil || container.Logger() == nil {
		t.Fatalf("container = %#v", container)
	}
}
