package createpage

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
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
	handler := NewCreatePageHandler(blobs, renderer, ids, clockStub{now: testCreatedAt}, index, testLogger())

	result, err := handler.Handle(context.Background(), CreatePageCommand{
		Title: " First page ",
		Area:  "Knowledge/Go",
		Tags:  []string{"Go", "Diseño de dominio"},
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.ID.String() != "page-id" || result.Path != "page/knowledge/go/first-page.adoc" {
		t.Fatalf("Handle() result = %+v", result)
	}
	if renderer.calls != 1 || renderer.name != "page" {
		t.Fatalf("renderer calls = %d, name = %q", renderer.calls, renderer.name)
	}
	page, ok := renderer.data.(domain.Page)
	if !ok || page.Title().String() != "First page" || page.Area().String() != "knowledge/go" || page.Tags().Len() != 2 {
		t.Fatalf("Render() data = %#v", renderer.data)
	}
	if blobs.createCalls != 1 || blobs.created.Path != result.Path || string(blobs.created.Content) != "rendered page" {
		t.Fatalf("created blob = %+v, calls = %d", blobs.created, blobs.createCalls)
	}
	if index.calls != 1 || index.document.Path != result.Path || index.document.ID != result.ID {
		t.Fatalf("indexed document = %+v, calls = %d", index.document, index.calls)
	}
}

func TestHandlerTranslatesExistingBlobToPageConflict(t *testing.T) {
	blobs := &blobWriterStub{createErr: ports.ErrBlobAlreadyExists}
	handler := newTestHandler(blobs, &rendererStub{}, &idGeneratorStub{id: "page-id"}, &indexWriterStub{}, testLogger())

	_, err := handler.Handle(context.Background(), CreatePageCommand{Title: "Page", Area: "Knowledge"})
	if !errors.Is(err, ErrPageAlreadyExists) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrPageAlreadyExists)
	}
	conflict, ok := errors.AsType[*PageAlreadyExistsError](err)
	if !ok || conflict.Path != "page/knowledge/page.adoc" {
		t.Fatalf("Handle() error = %#v", err)
	}
	if errors.Is(err, ports.ErrBlobAlreadyExists) {
		t.Fatalf("storage error escaped the use case: %v", err)
	}
}

func TestHandlerStopsWhenRenderingFails(t *testing.T) {
	renderErr := errors.New("template unavailable")
	blobs := &blobWriterStub{}
	index := &indexWriterStub{}
	handler := newTestHandler(blobs, &rendererStub{err: renderErr}, &idGeneratorStub{id: "page-id"}, index, testLogger())

	_, err := handler.Handle(context.Background(), CreatePageCommand{Title: "Page"})
	if !errors.Is(err, renderErr) || !strings.Contains(err.Error(), `render page "page/page.adoc"`) {
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
	handler := newTestHandler(blobs, &rendererStub{}, &idGeneratorStub{id: "page-id"}, index, testLogger())

	_, err := handler.Handle(context.Background(), CreatePageCommand{Title: "Page"})
	if !errors.Is(err, storageErr) || !strings.Contains(err.Error(), `store page "page/page.adoc"`) {
		t.Fatalf("Handle() error = %v", err)
	}
	if index.calls != 0 {
		t.Fatalf("index calls = %d, want 0", index.calls)
	}
}

func TestHandlerRollsBackStoredPageWhenIndexingFails(t *testing.T) {
	indexErr := errors.New("index unavailable")
	blobs := &blobWriterStub{}
	handler := newTestHandler(blobs, &rendererStub{}, &idGeneratorStub{id: "page-id"}, &indexWriterStub{err: indexErr}, testLogger())

	_, err := handler.Handle(context.Background(), CreatePageCommand{Title: "Page"})
	if !errors.Is(err, indexErr) || !strings.Contains(err.Error(), `index page "page/page.adoc"`) {
		t.Fatalf("Handle() error = %v", err)
	}
	if blobs.deleteCalls != 1 || blobs.deletedPath != "page/page.adoc" {
		t.Fatalf("Delete() path = %q, calls = %d", blobs.deletedPath, blobs.deleteCalls)
	}
}

func TestHandlerReportsIncompleteCreationWhenRollbackFails(t *testing.T) {
	indexErr := errors.New("index unavailable")
	cleanupErr := errors.New("delete unavailable")
	blobs := &blobWriterStub{deleteErr: cleanupErr}
	handler := newTestHandler(blobs, &rendererStub{}, &idGeneratorStub{id: "page-id"}, &indexWriterStub{err: indexErr}, testLogger())

	_, err := handler.Handle(context.Background(), CreatePageCommand{Title: "Page"})
	if !errors.Is(err, ErrPageCreationIncomplete) || !errors.Is(err, indexErr) || !errors.Is(err, cleanupErr) {
		t.Fatalf("Handle() error = %v", err)
	}
	incomplete, ok := errors.AsType[*PageCreationIncompleteError](err)
	if !ok || incomplete.Path != "page/page.adoc" {
		t.Fatalf("Handle() error = %#v", err)
	}
}

func TestHandlerLogsExpectedAndInternalFailuresAtDebugLevel(t *testing.T) {
	tests := map[string]struct {
		command CreatePageCommand
		ids     *idGeneratorStub
		render  *rendererStub
		blobs   *blobWriterStub
		want    string
	}{
		"validation":     {command: CreatePageCommand{}, ids: &idGeneratorStub{id: "page-id"}, render: &rendererStub{}, blobs: &blobWriterStub{}, want: "page validation failed"},
		"initialization": {command: CreatePageCommand{Title: "Page"}, ids: &idGeneratorStub{err: errors.New("random source unavailable")}, render: &rendererStub{}, blobs: &blobWriterStub{}, want: "page initialization failed"},
		"rendering":      {command: CreatePageCommand{Title: "Page"}, ids: &idGeneratorStub{id: "page-id"}, render: &rendererStub{err: errors.New("template unavailable")}, blobs: &blobWriterStub{}, want: "page rendering failed"},
		"storage":        {command: CreatePageCommand{Title: "Page"}, ids: &idGeneratorStub{id: "page-id"}, render: &rendererStub{}, blobs: &blobWriterStub{createErr: errors.New("disk unavailable")}, want: "page storage failed"},
		"existing":       {command: CreatePageCommand{Title: "Page"}, ids: &idGeneratorStub{id: "page-id"}, render: &rendererStub{}, blobs: &blobWriterStub{createErr: ports.ErrBlobAlreadyExists}, want: "page already exists"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
			handler := newTestHandler(test.blobs, test.render, test.ids, &indexWriterStub{}, logger)

			_, _ = handler.Handle(context.Background(), test.command)
			if !strings.Contains(output.String(), `"level":"DEBUG"`) || !strings.Contains(output.String(), `"msg":"`+test.want+`"`) {
				t.Fatalf("logs = %s", output.String())
			}
		})
	}
}

func TestNewCreatePageHandlerRequiresDependencies(t *testing.T) {
	validWriter := &blobWriterStub{}
	validRenderer := &rendererStub{}
	validIDs := &idGeneratorStub{id: "page-id"}
	validClock := clockStub{now: testCreatedAt}
	validIndex := &indexWriterStub{}
	validLogger := testLogger()

	for name, build := range map[string]func(){
		"blob writer":    func() { NewCreatePageHandler(nil, validRenderer, validIDs, validClock, validIndex, validLogger) },
		"renderer":       func() { NewCreatePageHandler(validWriter, nil, validIDs, validClock, validIndex, validLogger) },
		"id generator":   func() { NewCreatePageHandler(validWriter, validRenderer, nil, validClock, validIndex, validLogger) },
		"clock":          func() { NewCreatePageHandler(validWriter, validRenderer, validIDs, nil, validIndex, validLogger) },
		"document index": func() { NewCreatePageHandler(validWriter, validRenderer, validIDs, validClock, nil, validLogger) },
		"logger":         func() { NewCreatePageHandler(validWriter, validRenderer, validIDs, validClock, validIndex, nil) },
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
	logger *slog.Logger,
) *CreatePageHandler {
	return NewCreatePageHandler(blobs, renderer, ids, clockStub{now: testCreatedAt}, index, logger)
}
