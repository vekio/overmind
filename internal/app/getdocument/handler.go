package getdocument

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

// GetDocumentHandler retrieves documents from the query-side index.
type GetDocumentHandler struct {
	index ports.DocumentIndexReader
}

// NewGetDocumentHandler creates the query handler.
func NewGetDocumentHandler(index ports.DocumentIndexReader) *GetDocumentHandler {
	if index == nil {
		panic("get document handler requires document index reader")
	}
	return &GetDocumentHandler{index: index}
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
	return GetDocumentResult{
		ID:         document.ID,
		Path:       document.Path,
		Content:    append([]byte(nil), document.Content...),
		Attributes: cloneAttributes(document.Attributes),
	}, nil
}

func cloneAttributes(attributes map[string]string) map[string]string {
	result := make(map[string]string, len(attributes))
	for name, value := range attributes {
		result[name] = value
	}
	return result
}
