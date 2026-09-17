package createpage

import (
	"context"
	"errors"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

func TestHandlerCreatesPage(t *testing.T) {
	blobs := &blobWriterStub{}
	renderer := &rendererStub{}
	ids := &idGeneratorStub{id: "page-id"}
	index := &indexWriterStub{}
	handler := NewCreatePageHandler(blobs, renderer, ids, clockStub{now: testCreatedAt}, index)

	result, err := handler.Handle(context.Background(), CreatePageCommand{
		Title: " First page ",
		Area:  "Knowledge/Go",
		Tags:  []string{"Go", "Diseño de dominio"},
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.ID.String() != "page-id" {
		t.Fatalf("Handle() result = %+v", result)
	}
	if renderer.calls != 1 || renderer.name != "page" {
		t.Fatalf("renderer calls = %d, name = %q", renderer.calls, renderer.name)
	}
	page, ok := renderer.data.(domain.Page)
	if !ok || page.Title().String() != "First page" || page.Area().String() != "knowledge/go" || page.Tags().Len() != 2 {
		t.Fatalf("Render() data = %#v", renderer.data)
	}
	if blobs.createCalls != 1 || blobs.created.ID != result.ID || string(blobs.created.Content) != "rendered page" {
		t.Fatalf("created blob = %+v, calls = %d", blobs.created, blobs.createCalls)
	}
	if index.calls != 1 || index.document.ID != result.ID || index.document.Area != "knowledge/go" {
		t.Fatalf("indexed document = %+v, calls = %d", index.document, index.calls)
	}
}

func TestHandlerPreservesExistingBlobError(t *testing.T) {
	blobs := &blobWriterStub{createErr: ports.ErrBlobAlreadyExists}
	handler := newTestHandler(blobs, &rendererStub{}, &idGeneratorStub{id: "page-id"}, &indexWriterStub{})

	_, err := handler.Handle(context.Background(), CreatePageCommand{Title: "Page", Area: "Knowledge"})
	if !errors.Is(err, ports.ErrBlobAlreadyExists) ||
		!strings.Contains(err.Error(), `store page "page-id"`) {
		t.Fatalf("Handle() error = %v", err)
	}
}

func TestHandlerStopsWhenRenderingFails(t *testing.T) {
	renderErr := errors.New("template unavailable")
	blobs := &blobWriterStub{}
	index := &indexWriterStub{}
	handler := newTestHandler(blobs, &rendererStub{err: renderErr}, &idGeneratorStub{id: "page-id"}, index)

	_, err := handler.Handle(context.Background(), CreatePageCommand{Title: "Page"})
	if !errors.Is(err, renderErr) || !strings.Contains(err.Error(), `render page "page-id"`) {
		t.Fatalf("Handle() error = %v", err)
	}
	if blobs.createCalls != 0 || index.calls != 0 {
		t.Fatalf("blob calls = %d, index calls = %d", blobs.createCalls, index.calls)
	}
}

func TestHandlerStopsWhenStorageFails(t *testing.T) {
	storageErr := errors.New("disk unavailable")
	blobs := &blobWriterStub{createErr: storageErr}
	index := &indexWriterStub{}
	handler := newTestHandler(blobs, &rendererStub{}, &idGeneratorStub{id: "page-id"}, index)

	_, err := handler.Handle(context.Background(), CreatePageCommand{Title: "Page"})
	if !errors.Is(err, storageErr) || !strings.Contains(err.Error(), `store page "page-id"`) {
		t.Fatalf("Handle() error = %v", err)
	}
	if index.calls != 0 {
		t.Fatalf("index calls = %d, want 0", index.calls)
	}
}

func TestHandlerRollsBackStoredPageWhenIndexingFails(t *testing.T) {
	indexErr := errors.New("index unavailable")
	blobs := &blobWriterStub{}
	handler := newTestHandler(blobs, &rendererStub{}, &idGeneratorStub{id: "page-id"}, &indexWriterStub{err: indexErr})

	_, err := handler.Handle(context.Background(), CreatePageCommand{Title: "Page"})
	if !errors.Is(err, indexErr) || !strings.Contains(err.Error(), `index page "page-id"`) {
		t.Fatalf("Handle() error = %v", err)
	}
	if blobs.deleteCalls != 1 || blobs.deletedID.String() != "page-id" {
		t.Fatalf("Delete() ID = %q, calls = %d", blobs.deletedID, blobs.deleteCalls)
	}
}

func TestHandlerPreservesIndexAndRollbackFailures(t *testing.T) {
	indexErr := errors.New("index unavailable")
	cleanupErr := errors.New("delete unavailable")
	blobs := &blobWriterStub{deleteErr: cleanupErr}
	handler := newTestHandler(blobs, &rendererStub{}, &idGeneratorStub{id: "page-id"}, &indexWriterStub{err: indexErr})

	_, err := handler.Handle(context.Background(), CreatePageCommand{Title: "Page"})
	if !errors.Is(err, indexErr) || !errors.Is(err, cleanupErr) ||
		!strings.Contains(err.Error(), `index page "page-id"`) ||
		!strings.Contains(err.Error(), "remove created page") {
		t.Fatalf("Handle() error = %v", err)
	}
}

func TestNewCreatePageHandlerRequiresDependencies(t *testing.T) {
	validWriter := &blobWriterStub{}
	validRenderer := &rendererStub{}
	validIDs := &idGeneratorStub{id: "page-id"}
	validClock := clockStub{now: testCreatedAt}
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

func newTestHandler(
	blobs ports.BlobWriter,
	renderer ports.Renderer,
	ids ports.IDGenerator,
	index ports.DocumentIndexWriter,
) *CreatePageHandler {
	return NewCreatePageHandler(blobs, renderer, ids, clockStub{now: testCreatedAt}, index)
}
