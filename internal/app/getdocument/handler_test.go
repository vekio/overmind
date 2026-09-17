package getdocument

import (
	"context"
	"errors"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

type blobReaderStub struct {
	id   domain.DocumentID
	blob ports.Blob
	err  error
}

func (reader *blobReaderStub) Get(_ context.Context, id domain.DocumentID) (ports.Blob, error) {
	reader.id = id
	return reader.blob, reader.err
}

func (*blobReaderStub) List(context.Context) ([]domain.DocumentID, error) {
	return nil, nil
}

func TestHandlerGetsRawDocumentByID(t *testing.T) {
	id, _ := domain.NewDocumentID("page-id")
	blobs := &blobReaderStub{blob: ports.Blob{
		ID:       id,
		Content:  []byte("= Page\n"),
		Revision: "revision-1",
	}}
	handler := NewGetDocumentHandler(blobs)

	result, err := handler.Handle(context.Background(), GetDocumentQuery{ID: id})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if blobs.id != id || result.ID != blobs.blob.ID ||
		string(result.Content) != "= Page\n" || result.Revision != "revision-1" {
		t.Fatalf("ID = %q, result = %+v", blobs.id, result)
	}

	blobs.blob.Content[0] = '!'
	if string(result.Content) != "= Page\n" {
		t.Fatal("result content aliases blob content")
	}
}

func TestHandlerPreservesBlobFailure(t *testing.T) {
	blobErr := errors.New("blob not found")
	id, _ := domain.NewDocumentID("missing-id")
	_, err := NewGetDocumentHandler(&blobReaderStub{err: blobErr}).Handle(
		context.Background(),
		GetDocumentQuery{ID: id},
	)
	if !errors.Is(err, blobErr) || !strings.Contains(err.Error(), `get document "missing-id"`) {
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
