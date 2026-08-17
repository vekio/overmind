package updatedocument

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"git.casta.me/alberto/overmind/internal/ports"
)

var testUpdatedAt = time.Date(2026, time.August, 17, 12, 30, 0, 0, time.UTC)

type clockStub struct{ now time.Time }

func (clock clockStub) Now() time.Time { return clock.now }

type blobReaderStub struct {
	blob ports.Blob
	err  error
}

func (reader *blobReaderStub) Get(context.Context, string) (ports.Blob, error) {
	return reader.blob, reader.err
}

func (*blobReaderStub) List(context.Context, ports.BlobFilter) ([]string, error) { return nil, nil }

type blobUpdaterStub struct {
	updates   []ports.Blob
	expected  []string
	revisions []string
	errors    []error
}

func (updater *blobUpdaterStub) Update(_ context.Context, blob ports.Blob, expected string) (string, error) {
	updater.updates = append(updater.updates, blob)
	updater.expected = append(updater.expected, expected)
	index := len(updater.updates) - 1
	var revision string
	if index < len(updater.revisions) {
		revision = updater.revisions[index]
	}
	var err error
	if index < len(updater.errors) {
		err = updater.errors[index]
	}
	return revision, err
}

type indexWriterStub struct {
	document ports.IndexedDocument
	calls    int
	err      error
}

func (index *indexWriterStub) Upsert(_ context.Context, document ports.IndexedDocument) error {
	index.document = document
	index.calls++
	return index.err
}

func (*indexWriterStub) ReplaceAll(context.Context, []ports.IndexedDocument) error { return nil }

func TestHandlerUpdatesAndIndexesDocument(t *testing.T) {
	reader := &blobReaderStub{blob: ports.Blob{
		Path: "page/page.adoc", Content: source("Page", "page-id", "page", "Body"), Revision: "old-revision",
	}}
	updater := &blobUpdaterStub{revisions: []string{"new-revision"}}
	index := &indexWriterStub{}
	handler := newTestHandler(reader, updater, index)

	result, err := handler.Handle(context.Background(), UpdateDocumentCommand{
		Path: "page/page.adoc", Content: source("Updated title", "page-id", "page", "Updated body"), ExpectedRevision: "old-revision",
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.Path != "page/page.adoc" || result.Revision != "new-revision" || len(updater.updates) != 1 ||
		updater.expected[0] != "old-revision" || index.calls != 1 || index.document.Title != "Updated title" ||
		!index.document.UpdatedAt.Equal(testUpdatedAt) ||
		!strings.Contains(string(updater.updates[0].Content), ":overmind-updated-at: 2026-08-17T12:30:00Z") {
		t.Fatalf("result = %+v, updates = %+v, expected = %v, indexed = %+v", result, updater.updates, updater.expected, index.document)
	}
}

func TestHandlerReturnsWithoutWritingUnchangedContent(t *testing.T) {
	current := source("Page", "page-id", "page", "Body")
	reader := &blobReaderStub{blob: ports.Blob{Path: "page/page.adoc", Content: current, Revision: "revision"}}
	updater := &blobUpdaterStub{}
	index := &indexWriterStub{}
	result, err := newTestHandler(reader, updater, index).Handle(context.Background(), UpdateDocumentCommand{
		Path: "page/page.adoc", Content: append([]byte(nil), current...), ExpectedRevision: "revision",
	})
	if err != nil || result.Revision != "revision" || len(updater.updates) != 0 || index.calls != 0 {
		t.Fatalf("Handle() = (%+v, %v), updates = %d, index calls = %d", result, err, len(updater.updates), index.calls)
	}
}

func TestHandlerRejectsStaleRevision(t *testing.T) {
	reader := &blobReaderStub{blob: ports.Blob{Path: "page/page.adoc", Revision: "current"}}
	_, err := newTestHandler(reader, &blobUpdaterStub{}, &indexWriterStub{}).Handle(
		context.Background(), UpdateDocumentCommand{Path: "page/page.adoc", ExpectedRevision: "stale"},
	)
	if !errors.Is(err, ports.ErrBlobChanged) {
		t.Fatalf("Handle() error = %v, want %v", err, ports.ErrBlobChanged)
	}
}

func TestHandlerPreservesReadAndStorageFailures(t *testing.T) {
	readErr := errors.New("read unavailable")
	_, err := newTestHandler(
		&blobReaderStub{err: readErr}, &blobUpdaterStub{}, &indexWriterStub{},
	).Handle(context.Background(), UpdateDocumentCommand{Path: "page/page.adoc", ExpectedRevision: "revision"})
	if !errors.Is(err, readErr) || !strings.Contains(err.Error(), "read current content") {
		t.Fatalf("Handle(read) error = %v", err)
	}

	storageErr := errors.New("write unavailable")
	current := ports.Blob{Path: "page/page.adoc", Content: source("Page", "page-id", "page", "Body"), Revision: "revision"}
	updater := &blobUpdaterStub{errors: []error{storageErr}}
	index := &indexWriterStub{}
	_, err = newTestHandler(&blobReaderStub{blob: current}, updater, index).Handle(
		context.Background(), UpdateDocumentCommand{
			Path: current.Path, Content: source("Updated", "page-id", "page", "Body"), ExpectedRevision: current.Revision,
		},
	)
	if !errors.Is(err, storageErr) || !strings.Contains(err.Error(), "store edited content") || index.calls != 0 {
		t.Fatalf("Handle(storage) error = %v, index calls = %d", err, index.calls)
	}
}

func TestHandlerRejectsInvalidOrChangedMetadata(t *testing.T) {
	current := ports.Blob{Path: "page/page.adoc", Content: source("Page", "page-id", "page", "Body"), Revision: "revision"}
	for name, content := range map[string][]byte{
		"unmanaged":        []byte("= External\n\nBody\n"),
		"changed id":       source("Page", "other-id", "page", "Body"),
		"changed kind":     source("Page", "page-id", "unknown", "Body"),
		"invalid AsciiDoc": append(source("Page", "page-id", "page", "Body"), []byte("\n= Late title\n")...),
	} {
		t.Run(name, func(t *testing.T) {
			updater := &blobUpdaterStub{}
			_, err := newTestHandler(&blobReaderStub{blob: current}, updater, &indexWriterStub{}).Handle(
				context.Background(), UpdateDocumentCommand{Path: current.Path, Content: content, ExpectedRevision: current.Revision},
			)
			if err == nil || len(updater.updates) != 0 {
				t.Fatalf("Handle() error = %v, updates = %d", err, len(updater.updates))
			}
		})
	}
}

func TestHandlerRollsBackBlobWhenIndexingFails(t *testing.T) {
	indexErr := errors.New("index unavailable")
	current := ports.Blob{Path: "page/page.adoc", Content: source("Page", "page-id", "page", "Body"), Revision: "old-revision"}
	updater := &blobUpdaterStub{revisions: []string{"new-revision", "restored-revision"}}
	_, err := newTestHandler(&blobReaderStub{blob: current}, updater, &indexWriterStub{err: indexErr}).Handle(
		context.Background(), UpdateDocumentCommand{
			Path: current.Path, Content: source("Updated", "page-id", "page", "Body"), ExpectedRevision: current.Revision,
		},
	)
	if !errors.Is(err, indexErr) || !strings.Contains(err.Error(), "index edited content") || len(updater.updates) != 2 ||
		updater.expected[1] != "new-revision" || string(updater.updates[1].Content) != string(current.Content) {
		t.Fatalf("Handle() error = %v, updates = %+v, expected = %v", err, updater.updates, updater.expected)
	}
}

func TestHandlerPreservesIndexAndRollbackFailures(t *testing.T) {
	indexErr := errors.New("index unavailable")
	rollbackErr := errors.New("rollback unavailable")
	current := ports.Blob{Path: "page/page.adoc", Content: source("Page", "page-id", "page", "Body"), Revision: "old-revision"}
	updater := &blobUpdaterStub{revisions: []string{"new-revision"}, errors: []error{nil, rollbackErr}}
	_, err := newTestHandler(&blobReaderStub{blob: current}, updater, &indexWriterStub{err: indexErr}).Handle(
		context.Background(), UpdateDocumentCommand{
			Path: current.Path, Content: source("Updated", "page-id", "page", "Body"), ExpectedRevision: current.Revision,
		},
	)
	if !errors.Is(err, indexErr) || !errors.Is(err, rollbackErr) || !strings.Contains(err.Error(), "restore previous content") {
		t.Fatalf("Handle() error = %v", err)
	}
}

func TestHandlerRequiresRevision(t *testing.T) {
	_, err := newTestHandler(&blobReaderStub{}, &blobUpdaterStub{}, &indexWriterStub{}).Handle(
		context.Background(), UpdateDocumentCommand{},
	)
	if !errors.Is(err, ErrRevisionRequired) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrRevisionRequired)
	}
}

func TestNewHandlerRequiresDependencies(t *testing.T) {
	reader := &blobReaderStub{}
	updater := &blobUpdaterStub{}
	clock := clockStub{now: testUpdatedAt}
	index := &indexWriterStub{}
	for name, build := range map[string]func(){
		"blob reader":    func() { NewUpdateDocumentHandler(nil, updater, clock, index) },
		"blob updater":   func() { NewUpdateDocumentHandler(reader, nil, clock, index) },
		"clock":          func() { NewUpdateDocumentHandler(reader, updater, nil, index) },
		"document index": func() { NewUpdateDocumentHandler(reader, updater, clock, nil) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("NewUpdateDocumentHandler() did not panic")
				}
			}()
			build()
		})
	}
}

func newTestHandler(
	reader ports.BlobReader,
	updater ports.BlobUpdater,
	index ports.DocumentIndexWriter,
) *UpdateDocumentHandler {
	return NewUpdateDocumentHandler(reader, updater, clockStub{now: testUpdatedAt}, index)
}

func source(title, id, kind, body string) []byte {
	return []byte("= " + title + "\n" +
		":overmind-id: " + id + "\n" +
		":overmind-type: " + kind + "\n" +
		":overmind-created-at: 2026-08-16T10:00:00Z\n" +
		":overmind-updated-at: 2026-08-16T10:00:00Z\n\n" + body + "\n")
}
