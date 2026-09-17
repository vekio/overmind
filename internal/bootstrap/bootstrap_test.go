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

func TestApplicationCreatesAnAsciiDocPage(t *testing.T) {
	root := t.TempDir()
	application, err := New(config.Config{DataDir: root})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() {
		if err := application.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	result, err := application.Commands.CreatePage.Handle(context.Background(), createpage.CreatePageCommand{
		Title: "First page",
		Area:  "Knowledge",
	})
	if err != nil {
		t.Fatalf("CreatePage.Handle() error = %v", err)
	}
	if result.ID.String() == "" {
		t.Fatal("CreatePage.Handle() returned an empty ID")
	}
	content, err := os.ReadFile(filepath.Join(root, "documents", result.ID.String()+".adoc"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := string(content); !strings.HasPrefix(got, "= First page\n:overmind-id: ") {
		t.Fatalf("page content = %q", got)
	}
	retrieved, err := application.Queries.GetDocument.Handle(context.Background(), getdocument.GetDocumentQuery{
		ID: result.ID,
	})
	if err != nil {
		t.Fatalf("GetDocument.Handle() error = %v", err)
	}
	if retrieved.ID != result.ID || string(retrieved.Content) != string(content) || retrieved.Revision == "" {
		t.Fatalf("retrieved document = %+v", retrieved)
	}
	editedContent := []byte(strings.Replace(string(retrieved.Content), "= First page", "= Updated page", 1))
	updated, err := application.Commands.UpdateDocument.Handle(context.Background(), updatedocument.UpdateDocumentCommand{
		ID:               retrieved.ID,
		Content:          editedContent,
		ExpectedRevision: retrieved.Revision,
	})
	if err != nil {
		t.Fatalf("UpdateDocument.Handle() error = %v", err)
	}
	if updated.ID != retrieved.ID || updated.Revision == "" || updated.Revision == retrieved.Revision {
		t.Fatalf("updated document = %+v", updated)
	}
	storedContent, err := os.ReadFile(filepath.Join(root, "documents", result.ID.String()+".adoc"))
	if err != nil || string(storedContent) != string(editedContent) {
		t.Fatalf("stored edited content = %q, error = %v", storedContent, err)
	}
	listed, err := application.Queries.ListDocuments.Handle(context.Background(), listdocuments.ListDocumentsQuery{
		Type: "page",
	})
	if err != nil {
		t.Fatalf("ListDocuments.Handle() error = %v", err)
	}
	if len(listed.Documents) != 1 || listed.Documents[0].ID != result.ID || listed.Documents[0].Title != "Updated page" {
		t.Fatalf("listed documents = %+v", listed.Documents)
	}
}
