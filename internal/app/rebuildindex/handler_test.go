package rebuildindex

import (
	"context"
	"errors"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

func validDocumentSource() []byte {
	return []byte("= Page\n" +
		":overmind-id: page-id\n" +
		":overmind-type: page\n" +
		":overmind-area: knowledge\n" +
		":overmind-tags: Go\n" +
		":overmind-created-at: 2026-08-16T10:00:00Z\n" +
		":overmind-updated-at: 2026-08-16T10:00:00Z\n")
}

type blobReaderStub struct {
	ids       []domain.DocumentID
	blobs     map[string]ports.Blob
	getErrors map[string]error
	listErr   error
}

func (reader *blobReaderStub) List(context.Context) ([]domain.DocumentID, error) {
	return reader.ids, reader.listErr
}

func (reader *blobReaderStub) Get(_ context.Context, id domain.DocumentID) (ports.Blob, error) {
	if err := reader.getErrors[id.String()]; err != nil {
		return ports.Blob{}, err
	}
	return reader.blobs[id.String()], nil
}

type indexWriterStub struct {
	documents    []ports.IndexedDocument
	replaceCalls int
	err          error
}

func (*indexWriterStub) Upsert(context.Context, ports.IndexedDocument) error { return nil }
func (index *indexWriterStub) ReplaceAll(_ context.Context, documents []ports.IndexedDocument) error {
	index.replaceCalls++
	index.documents = documents
	return index.err
}

func TestHandlerRebuildsIndexFromDocumentsWhoseIDMatchesStorage(t *testing.T) {
	id, _ := domain.NewDocumentID("page-id")
	blobs := &blobReaderStub{
		ids:   []domain.DocumentID{id},
		blobs: map[string]ports.Blob{id.String(): {ID: id, Content: validDocumentSource()}},
	}
	index := &indexWriterStub{}
	result, err := NewRebuildIndexHandler(blobs, index).Handle(context.Background(), RebuildIndexCommand{})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.Documents != 1 || index.replaceCalls != 1 || len(index.documents) != 1 ||
		index.documents[0].ID != id || index.documents[0].Area != "knowledge" {
		t.Fatalf("result = %+v, indexed = %+v", result, index.documents)
	}
}

func TestHandlerRejectsEmbeddedIDMismatch(t *testing.T) {
	storedID, _ := domain.NewDocumentID("other-id")
	blobs := &blobReaderStub{
		ids:   []domain.DocumentID{storedID},
		blobs: map[string]ports.Blob{storedID.String(): {ID: storedID, Content: validDocumentSource()}},
	}
	_, err := NewRebuildIndexHandler(blobs, &indexWriterStub{}).Handle(context.Background(), RebuildIndexCommand{})
	if err == nil || !strings.Contains(err.Error(), "contains ID") {
		t.Fatalf("Handle() error = %v", err)
	}
}

func TestHandlerPreservesStorageAndIndexFailures(t *testing.T) {
	listErr := errors.New("storage unavailable")
	index := &indexWriterStub{}
	_, err := NewRebuildIndexHandler(&blobReaderStub{listErr: listErr}, index).Handle(context.Background(), RebuildIndexCommand{})
	if !errors.Is(err, listErr) || index.replaceCalls != 0 {
		t.Fatalf("Handle() error = %v, calls = %d", err, index.replaceCalls)
	}

	replaceErr := errors.New("database unavailable")
	index = &indexWriterStub{err: replaceErr}
	_, err = NewRebuildIndexHandler(&blobReaderStub{}, index).Handle(context.Background(), RebuildIndexCommand{})
	if !errors.Is(err, replaceErr) || index.replaceCalls != 1 {
		t.Fatalf("Handle() error = %v, calls = %d", err, index.replaceCalls)
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
					t.Fatal("constructor did not panic")
				}
			}()
			build()
		})
	}
}
