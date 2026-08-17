package getdocument

import (
	"context"
	"errors"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/ports"
)

type blobReaderStub struct {
	path string
	blob ports.Blob
	err  error
}

func (reader *blobReaderStub) Get(_ context.Context, path string) (ports.Blob, error) {
	reader.path = path
	return reader.blob, reader.err
}

func (*blobReaderStub) List(context.Context, ports.BlobFilter) ([]string, error) {
	return nil, nil
}

func TestHandlerGetsRawDocumentByPath(t *testing.T) {
	blobs := &blobReaderStub{blob: ports.Blob{
		Path:     "page/knowledge/page.adoc",
		Content:  []byte("= Page\n"),
		Revision: "revision-1",
	}}
	handler := NewGetDocumentHandler(blobs)

	result, err := handler.Handle(context.Background(), GetDocumentQuery{Path: "page/knowledge/page.adoc"})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if blobs.path != "page/knowledge/page.adoc" || result.Path != blobs.blob.Path ||
		string(result.Content) != "= Page\n" || result.Revision != "revision-1" {
		t.Fatalf("path = %q, result = %+v", blobs.path, result)
	}

	blobs.blob.Content[0] = '!'
	if string(result.Content) != "= Page\n" {
		t.Fatal("result content aliases blob content")
	}
}

func TestHandlerPreservesBlobFailure(t *testing.T) {
	blobErr := errors.New("blob not found")
	_, err := NewGetDocumentHandler(&blobReaderStub{err: blobErr}).Handle(
		context.Background(),
		GetDocumentQuery{Path: "page/missing.adoc"},
	)
	if !errors.Is(err, blobErr) || !strings.Contains(err.Error(), `get document "page/missing.adoc"`) {
		t.Fatalf("Handle() error = %v", err)
	}
}

func TestNewHandlerRequiresBlobReader(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewGetDocumentHandler() did not panic")
		}
	}()
	NewGetDocumentHandler(nil)
}
