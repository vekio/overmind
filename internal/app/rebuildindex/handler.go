package rebuildindex

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/app/documentindex"
	"git.casta.me/alberto/overmind/internal/app/documentparser"
	"git.casta.me/alberto/overmind/internal/ports"
)

// RebuildIndexHandler reconstructs the read model from managed AsciiDoc blobs.
type RebuildIndexHandler struct {
	blobs ports.BlobReader
	index ports.DocumentIndexWriter
}

// NewRebuildIndexHandler creates the command handler.
func NewRebuildIndexHandler(blobs ports.BlobReader, index ports.DocumentIndexWriter) *RebuildIndexHandler {
	if blobs == nil {
		panic("rebuild index handler requires blob reader")
	}
	if index == nil {
		panic("rebuild index handler requires document index writer")
	}
	return &RebuildIndexHandler{blobs: blobs, index: index}
}

// Handle rebuilds the index atomically from every managed .adoc blob.
func (handler *RebuildIndexHandler) Handle(ctx context.Context, _ RebuildIndexCommand) (RebuildIndexResult, error) {
	ids, err := handler.blobs.List(ctx)
	if err != nil {
		return RebuildIndexResult{}, fmt.Errorf("rebuild index: list documents: %w", err)
	}

	documents := make([]ports.IndexedDocument, 0, len(ids))
	for _, id := range ids {
		blob, err := handler.blobs.Get(ctx, id)
		if err != nil {
			return RebuildIndexResult{}, fmt.Errorf("rebuild index: read %q: %w", id, err)
		}
		parsedDocument, managed, err := documentparser.Parse(blob.Content)
		if err != nil {
			return RebuildIndexResult{}, fmt.Errorf("rebuild index: parse %q: %w", id, err)
		}
		if managed {
			if parsedDocument.ID != id {
				return RebuildIndexResult{}, fmt.Errorf(
					"rebuild index: document %q contains ID %q", id, parsedDocument.ID,
				)
			}
			documents = append(documents, documentindex.FromParsed(parsedDocument))
		}
	}

	if err := handler.index.ReplaceAll(ctx, documents); err != nil {
		return RebuildIndexResult{}, fmt.Errorf("rebuild index: %w", err)
	}
	return RebuildIndexResult{Documents: len(documents)}, nil
}
