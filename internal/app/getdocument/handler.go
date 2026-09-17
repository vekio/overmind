package getdocument

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/ports"
)

// GetDocumentHandler retrieves documents from the source blob store.
type GetDocumentHandler struct {
	blobs ports.BlobReader
}

// NewGetDocumentHandler creates the query handler.
func NewGetDocumentHandler(blobs ports.BlobReader) *GetDocumentHandler {
	if blobs == nil {
		panic("get document handler requires blob reader")
	}
	return &GetDocumentHandler{blobs: blobs}
}

// Handle retrieves the raw document with the requested ID.
func (handler *GetDocumentHandler) Handle(ctx context.Context, query GetDocumentQuery) (GetDocumentResult, error) {
	blob, err := handler.blobs.Get(ctx, query.ID)
	if err != nil {
		return GetDocumentResult{}, fmt.Errorf("get document %q: %w", query.ID, err)
	}
	return GetDocumentResult{
		ID:       blob.ID,
		Content:  append([]byte(nil), blob.Content...),
		Revision: blob.Revision,
	}, nil
}
