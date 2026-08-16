package getdocument

import (
	"context"
	"testing"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

type indexReaderStub struct {
	document ports.IndexedDocument
	err      error
}

type blobReaderStub struct {
	blob ports.Blob
	err  error
}

func (reader blobReaderStub) Get(context.Context, string) (ports.Blob, error) {
	return reader.blob, reader.err
}

func (blobReaderStub) List(context.Context, ports.BlobFilter) ([]string, error) {
	return nil, nil
}

func (index indexReaderStub) GetByID(context.Context, domain.DocumentID) (ports.IndexedDocument, error) {
	return index.document, index.err
}

func TestHandlerGetsDocumentByID(t *testing.T) {
	id, _ := domain.NewDocumentID("page-id")
	createdAt := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	handler := NewGetDocumentHandler(indexReaderStub{document: ports.IndexedDocument{
		ID:         id,
		Path:       "page/page.adoc",
		Kind:       "page",
		Title:      "Page",
		CreatedAt:  createdAt,
		Attributes: map[string]string{"area": "knowledge"},
	}}, blobReaderStub{blob: ports.Blob{Path: "page/page.adoc", Content: []byte("= Page\n")}})

	result, err := handler.Handle(context.Background(), GetDocumentQuery{ID: "page-id"})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.ID.String() != "page-id" || result.Path != "page/page.adoc" || result.Kind != "page" || result.Title != "Page" || !result.CreatedAt.Equal(createdAt) || string(result.Content) != "= Page\n" || result.Attributes["area"] != "knowledge" {
		t.Fatalf("Handle() = %+v", result)
	}
}

func TestNewHandlerRequiresDependencies(t *testing.T) {
	validIndex := indexReaderStub{}
	validBlobs := blobReaderStub{}
	for name, build := range map[string]func(){
		"index": func() { NewGetDocumentHandler(nil, validBlobs) },
		"blobs": func() { NewGetDocumentHandler(validIndex, nil) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("NewGetDocumentHandler() did not panic")
				}
			}()
			build()
		})
	}
}
