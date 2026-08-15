package getdocument

import (
	"context"
	"testing"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

type indexReaderStub struct {
	document ports.IndexedDocument
	err      error
}

func (index indexReaderStub) GetByID(context.Context, domain.DocumentID) (ports.IndexedDocument, error) {
	return index.document, index.err
}

func TestHandlerGetsDocumentByID(t *testing.T) {
	id, _ := domain.NewDocumentID("page-id")
	handler := NewGetDocumentHandler(indexReaderStub{document: ports.IndexedDocument{
		ID:         id,
		Path:       "page.adoc",
		Content:    []byte("= Page\n"),
		Attributes: map[string]string{domain.AttributeType: "page"},
	}})

	result, err := handler.Handle(context.Background(), GetDocumentQuery{ID: "page-id"})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.ID.String() != "page-id" || result.Path != "page.adoc" || result.Attributes[domain.AttributeType] != "page" {
		t.Fatalf("Handle() = %+v", result)
	}
}
