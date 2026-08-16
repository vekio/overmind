package createpage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/infrastructure/localfs"
	"git.casta.me/alberto/overmind/internal/ports"
)

type rendererStub struct {
	name string
	data any
}

func (renderer *rendererStub) Render(_ context.Context, name string, data any) ([]byte, error) {
	renderer.name = name
	renderer.data = data
	return []byte("rendered page"), nil
}

type idGeneratorStub struct{ id string }

func (generator idGeneratorStub) Generate() (string, error) { return generator.id, nil }

type countingIDGenerator struct {
	calls int
}

func (generator *countingIDGenerator) Generate() (string, error) {
	generator.calls++
	return "page-id", nil
}

type clockStub struct{ now time.Time }

func (clock clockStub) Now() time.Time { return clock.now }

var createdAt = time.Date(2026, time.August, 14, 10, 30, 0, 0, time.UTC)

type indexWriterStub struct {
	document ports.IndexedDocument
	err      error
}

func (index *indexWriterStub) Upsert(_ context.Context, document ports.IndexedDocument) error {
	index.document = document
	return index.err
}

func (index *indexWriterStub) ReplaceAll(context.Context, []ports.IndexedDocument) error {
	return nil
}

func TestHandlerRendersAndPersistsANewPage(t *testing.T) {
	root := t.TempDir()
	renderer := &rendererStub{}
	index := &indexWriterStub{}
	handler := NewCreatePageHandler(localfs.New(root), renderer, idGeneratorStub{id: "page-id"}, clockStub{now: createdAt}, index)

	result, err := handler.Handle(context.Background(), CreatePageCommand{
		Title: " First page ",
		Area:  "Knowledge/Go",
		Tags:  []string{"Go", "Diseño de dominio"},
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.ID.String() != "page-id" {
		t.Fatalf("Handle() ID = %q, want %q", result.ID, "page-id")
	}
	if renderer.name != "page" {
		t.Fatalf("Render() name = %q, want %q", renderer.name, "page")
	}
	page, ok := renderer.data.(domain.Page)
	if !ok || page.Title().String() != "First page" || page.Title().Slug() != "first-page" || page.Area().String() != "knowledge/go" || page.ID().String() != "page-id" || page.Tags().Len() != 2 || !page.CreatedAt().Equal(createdAt) {
		t.Fatalf("Render() data = %#v", renderer.data)
	}

	content, err := os.ReadFile(filepath.Join(root, "page", "knowledge", "go", "first-page.adoc"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "rendered page" {
		t.Fatalf("stored content = %q", content)
	}
	if index.document.ID.String() != "page-id" || index.document.Path != "page/knowledge/go/first-page.adoc" || index.document.Kind != domain.DocumentKindPage || index.document.Title != "First page" || len(index.document.Tags) != 2 || index.document.Attributes["area"] != "knowledge/go" {
		t.Fatalf("indexed document = %+v", index.document)
	}
}

func TestHandlerDoesNotReplaceExistingPage(t *testing.T) {
	store := localfs.New(t.TempDir())
	handler := NewCreatePageHandler(store, &rendererStub{}, idGeneratorStub{id: "page-id"}, clockStub{now: createdAt}, &indexWriterStub{})
	command := CreatePageCommand{Title: "Page", Area: "Knowledge"}

	if _, err := handler.Handle(context.Background(), command); err != nil {
		t.Fatalf("first Handle() error = %v", err)
	}
	if _, err := handler.Handle(context.Background(), command); !errors.Is(err, ports.ErrBlobAlreadyExists) {
		t.Fatalf("second Handle() error = %v, want %v", err, ports.ErrBlobAlreadyExists)
	}
}

func TestHandlerSavesPageWithoutAreaUnderTypeDirectory(t *testing.T) {
	for name, area := range map[string]string{
		"omitted":     "",
		"blank value": "   ",
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			handler := NewCreatePageHandler(localfs.New(root), &rendererStub{}, idGeneratorStub{id: "page-id"}, clockStub{now: createdAt}, &indexWriterStub{})

			if _, err := handler.Handle(context.Background(), CreatePageCommand{Title: "Root page", Area: area}); err != nil {
				t.Fatalf("Handle() error = %v", err)
			}
			if _, err := os.Stat(filepath.Join(root, "page", "root-page.adoc")); err != nil {
				t.Fatalf("Stat(page/root-page.adoc) error = %v", err)
			}
		})
	}
}

func TestHandlerRemovesBlobWhenIndexingFails(t *testing.T) {
	root := t.TempDir()
	handler := NewCreatePageHandler(
		localfs.New(root),
		&rendererStub{},
		idGeneratorStub{id: "page-id"},
		clockStub{now: createdAt},
		&indexWriterStub{err: errors.New("index unavailable")},
	)

	if _, err := handler.Handle(context.Background(), CreatePageCommand{Title: "Page"}); err == nil {
		t.Fatal("Handle() error = nil")
	}
	if _, err := os.Stat(filepath.Join(root, "page", "page.adoc")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat(page/page.adoc) error = %v, want not exist", err)
	}
}

func TestHandlerUsesDomainValidation(t *testing.T) {
	handler := NewCreatePageHandler(localfs.New(t.TempDir()), &rendererStub{}, idGeneratorStub{id: "page-id"}, clockStub{now: createdAt}, &indexWriterStub{})

	for name, command := range map[string]CreatePageCommand{
		"missing title": {Area: "Knowledge"},
		"invalid area":  {Title: "Page", Area: "///"},
		"invalid tag":   {Title: "Page", Tags: []string{"---"}},
		"duplicate tag": {Title: "Page", Tags: []string{"Go", "go"}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := handler.Handle(context.Background(), command); err == nil || !strings.Contains(err.Error(), "create page") {
				t.Fatalf("Handle() error = %v", err)
			}
		})
	}
}

func TestHandlerValidatesCommandBeforeGeneratingID(t *testing.T) {
	ids := &countingIDGenerator{}
	handler := NewCreatePageHandler(localfs.New(t.TempDir()), &rendererStub{}, ids, clockStub{now: createdAt}, &indexWriterStub{})

	if _, err := handler.Handle(context.Background(), CreatePageCommand{}); err == nil {
		t.Fatal("Handle() error = nil")
	}
	if ids.calls != 0 {
		t.Fatalf("Generate() calls = %d, want 0", ids.calls)
	}
}

type writerStub struct{}

func (writerStub) Create(context.Context, ports.Blob) error { return nil }
func (writerStub) Put(context.Context, ports.Blob) error    { return nil }
func (writerStub) Delete(context.Context, string) error     { return nil }

var _ ports.BlobWriter = writerStub{}

func TestNewCreatePageHandlerRequiresDependencies(t *testing.T) {
	validWriter := writerStub{}
	validRenderer := &rendererStub{}
	validIDs := idGeneratorStub{id: "page-id"}
	validClock := clockStub{now: createdAt}
	validIndex := &indexWriterStub{}

	for name, build := range map[string]func(){
		"blob writer":    func() { NewCreatePageHandler(nil, validRenderer, validIDs, validClock, validIndex) },
		"renderer":       func() { NewCreatePageHandler(validWriter, nil, validIDs, validClock, validIndex) },
		"id generator":   func() { NewCreatePageHandler(validWriter, validRenderer, nil, validClock, validIndex) },
		"clock":          func() { NewCreatePageHandler(validWriter, validRenderer, validIDs, nil, validIndex) },
		"document index": func() { NewCreatePageHandler(validWriter, validRenderer, validIDs, validClock, nil) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("NewCreatePageHandler() did not panic")
				}
			}()
			build()
		})
	}
}
