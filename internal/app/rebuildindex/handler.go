package rebuildindex

import (
	"context"
	"fmt"
	"strings"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
	"git.casta.me/alberto/overmind/pkg/asciidoc"
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
	paths, err := handler.blobs.List(ctx, ports.BlobFilter{Suffix: ".adoc"})
	if err != nil {
		return RebuildIndexResult{}, fmt.Errorf("rebuild index: list documents: %w", err)
	}

	documents := make([]ports.IndexedDocument, 0, len(paths))
	for _, path := range paths {
		blob, err := handler.blobs.Get(ctx, path)
		if err != nil {
			return RebuildIndexResult{}, fmt.Errorf("rebuild index: read %q: %w", path, err)
		}
		document, managed, err := indexedDocument(blob)
		if err != nil {
			return RebuildIndexResult{}, fmt.Errorf("rebuild index: parse %q: %w", path, err)
		}
		if managed {
			documents = append(documents, document)
		}
	}

	if err := handler.index.ReplaceAll(ctx, documents); err != nil {
		return RebuildIndexResult{}, fmt.Errorf("rebuild index: %w", err)
	}
	return RebuildIndexResult{Documents: len(documents)}, nil
}

func indexedDocument(blob ports.Blob) (ports.IndexedDocument, bool, error) {
	processed := asciidoc.Process(blob.Content)
	attributes := make(map[string]string)
	for _, attribute := range processed.Analysis.Header.Attributes.All() {
		if strings.HasPrefix(attribute.Name, domain.AttributePrefix) {
			attributes[attribute.Name] = attribute.Value
		}
	}
	idValue, managed := attributes[domain.AttributeID]
	if !managed {
		return ports.IndexedDocument{}, false, nil
	}
	if processed.HasErrors() {
		return ports.IndexedDocument{}, false, fmt.Errorf("invalid AsciiDoc: %s", processed.Diagnostics[0])
	}
	documentID, err := domain.NewDocumentID(idValue)
	if err != nil {
		return ports.IndexedDocument{}, false, err
	}

	return ports.IndexedDocument{
		ID:         documentID,
		Path:       blob.Path,
		Content:    append([]byte(nil), blob.Content...),
		Attributes: attributes,
	}, true, nil
}
