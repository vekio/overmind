package rebuildindex

import (
	"context"
	"errors"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/ports"
)

func validDocumentSource() []byte {
	return []byte("= Page\n" +
		":overmind-id: page-id\n" +
		":overmind-type: page\n" +
		":overmind-area: knowledge\n" +
		":overmind-tags: Go, Diseño de dominio\n" +
		":overmind-created-at: 2026-08-16T10:00:00Z\n" +
		":overmind-updated-at: 2026-08-16T10:00:00Z\n")
}

type blobReaderStub struct {
	paths      []string
	blobs      map[string]ports.Blob
	getErrors  map[string]error
	listFilter ports.BlobFilter
	listErr    error
}

func (reader *blobReaderStub) List(_ context.Context, filter ports.BlobFilter) ([]string, error) {
	reader.listFilter = filter
	return reader.paths, reader.listErr
}

func (reader *blobReaderStub) Get(_ context.Context, path string) (ports.Blob, error) {
	if err := reader.getErrors[path]; err != nil {
		return ports.Blob{}, err
	}
	return reader.blobs[path], nil
}

type indexWriterStub struct {
	documents    []ports.IndexedDocument
	replaceCalls int
	err          error
}

func (index *indexWriterStub) Upsert(context.Context, ports.IndexedDocument) error { return nil }

func (index *indexWriterStub) ReplaceAll(_ context.Context, documents []ports.IndexedDocument) error {
	index.replaceCalls++
	index.documents = documents
	return index.err
}

func TestHandlerRebuildsIndexFromManagedDocuments(t *testing.T) {
	blobs := &blobReaderStub{
		paths: []string{"page.adoc", "external.adoc"},
		blobs: map[string]ports.Blob{
			"page.adoc":     {Path: "page.adoc", Content: validDocumentSource()},
			"external.adoc": {Path: "external.adoc", Content: []byte("= External\n\nBody\n")},
		},
	}
	index := &indexWriterStub{}

	result, err := NewRebuildIndexHandler(blobs, index).Handle(context.Background(), RebuildIndexCommand{})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if blobs.listFilter.Suffix != ".adoc" {
		t.Fatalf("List() filter = %+v", blobs.listFilter)
	}
	if result.Documents != 1 || index.replaceCalls != 1 || len(index.documents) != 1 ||
		index.documents[0].Path != "page.adoc" {
		t.Fatalf("result = %+v, indexed = %+v, calls = %d", result, index.documents, index.replaceCalls)
	}
}

func TestHandlerStopsWhenBlobListingFails(t *testing.T) {
	listErr := errors.New("vault unavailable")
	index := &indexWriterStub{}
	_, err := NewRebuildIndexHandler(&blobReaderStub{listErr: listErr}, index).
		Handle(context.Background(), RebuildIndexCommand{})
	if !errors.Is(err, listErr) || !strings.Contains(err.Error(), "list documents") {
		t.Fatalf("Handle() error = %v", err)
	}
	if index.replaceCalls != 0 {
		t.Fatalf("ReplaceAll() calls = %d", index.replaceCalls)
	}
}

func TestHandlerStopsWhenBlobReadFails(t *testing.T) {
	readErr := errors.New("read unavailable")
	blobs := &blobReaderStub{
		paths:     []string{"page.adoc"},
		getErrors: map[string]error{"page.adoc": readErr},
	}
	index := &indexWriterStub{}
	_, err := NewRebuildIndexHandler(blobs, index).Handle(context.Background(), RebuildIndexCommand{})
	if !errors.Is(err, readErr) || !strings.Contains(err.Error(), `read "page.adoc"`) {
		t.Fatalf("Handle() error = %v", err)
	}
	if index.replaceCalls != 0 {
		t.Fatalf("ReplaceAll() calls = %d", index.replaceCalls)
	}
}

func TestHandlerStopsWhenManagedDocumentIsInvalid(t *testing.T) {
	blobs := &blobReaderStub{
		paths: []string{"page.adoc"},
		blobs: map[string]ports.Blob{"page.adoc": {
			Path: "page.adoc",
			Content: []byte("= Page\n:overmind-id: page-id\n:overmind-type: unknown\n" +
				":overmind-created-at: 2026-08-16T10:00:00Z\n\n"),
		}},
	}
	index := &indexWriterStub{}
	_, err := NewRebuildIndexHandler(blobs, index).Handle(context.Background(), RebuildIndexCommand{})
	if err == nil || !strings.Contains(err.Error(), `parse "page.adoc"`) ||
		!strings.Contains(err.Error(), "invalid overmind-type") {
		t.Fatalf("Handle() error = %v", err)
	}
	if index.replaceCalls != 0 {
		t.Fatalf("ReplaceAll() calls = %d", index.replaceCalls)
	}
}

func TestHandlerPreservesIndexReplacementFailure(t *testing.T) {
	replaceErr := errors.New("database unavailable")
	index := &indexWriterStub{err: replaceErr}
	_, err := NewRebuildIndexHandler(&blobReaderStub{}, index).
		Handle(context.Background(), RebuildIndexCommand{})
	if !errors.Is(err, replaceErr) || !strings.Contains(err.Error(), "rebuild index") {
		t.Fatalf("Handle() error = %v", err)
	}
	if index.replaceCalls != 1 {
		t.Fatalf("ReplaceAll() calls = %d", index.replaceCalls)
	}
}

func TestNewRebuildIndexHandlerRequiresDependencies(t *testing.T) {
	validBlobs := &blobReaderStub{}
	validIndex := &indexWriterStub{}
	for name, build := range map[string]func(){
		"blob reader":    func() { NewRebuildIndexHandler(nil, validIndex) },
		"document index": func() { NewRebuildIndexHandler(validBlobs, nil) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("NewRebuildIndexHandler() did not panic")
				}
			}()
			build()
		})
	}
}
