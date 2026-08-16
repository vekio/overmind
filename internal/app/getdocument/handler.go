package getdocument

import (
	"context"
	"fmt"
	"maps"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

// GetDocumentHandler locates documents through the index and reads their
// content from the source blob store.
type GetDocumentHandler struct {
	index ports.DocumentIndexReader
	blobs ports.BlobReader
}

// NewGetDocumentHandler creates the query handler.
func NewGetDocumentHandler(index ports.DocumentIndexReader, blobs ports.BlobReader) *GetDocumentHandler {
	if index == nil {
		panic("get document handler requires document index reader")
	}
	if blobs == nil {
		panic("get document handler requires blob reader")
	}
	return &GetDocumentHandler{index: index, blobs: blobs}
}

// Handle retrieves a document by its stable identifier.
func (handler *GetDocumentHandler) Handle(ctx context.Context, query GetDocumentQuery) (GetDocumentResult, error) {
	documentID, err := domain.NewDocumentID(query.ID)
	if err != nil {
		return GetDocumentResult{}, fmt.Errorf("get document: %w", err)
	}
	document, err := handler.index.GetByID(ctx, documentID)
	if err != nil {
		return GetDocumentResult{}, fmt.Errorf("get document %q: %w", documentID, err)
	}
	blob, err := handler.blobs.Get(ctx, document.Path)
	if err != nil {
		return GetDocumentResult{}, fmt.Errorf("get document %q content: %w", documentID, err)
	}
	return GetDocumentResult{
		ID:         document.ID,
		Path:       document.Path,
		Kind:       document.Kind,
		Title:      document.Title,
		CreatedAt:  document.CreatedAt,
		Content:    append([]byte(nil), blob.Content...),
		Attributes: cloneAttributes(document.Attributes),
	}, nil
}

func cloneAttributes(attributes map[string]string) map[string]string {
	result := make(map[string]string, len(attributes))
	maps.Copy(result, attributes)
	return result
}
