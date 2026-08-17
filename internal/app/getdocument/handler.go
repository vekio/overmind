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

// Handle retrieves the raw document at the requested logical path.
func (handler *GetDocumentHandler) Handle(ctx context.Context, query GetDocumentQuery) (GetDocumentResult, error) {
	blob, err := handler.blobs.Get(ctx, query.Path)
	if err != nil {
		return GetDocumentResult{}, fmt.Errorf("get document %q: %w", query.Path, err)
	}
	return GetDocumentResult{
		Path:     blob.Path,
		Content:  append([]byte(nil), blob.Content...),
		Revision: blob.Revision,
	}, nil
}
