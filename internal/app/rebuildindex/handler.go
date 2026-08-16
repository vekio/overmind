package rebuildindex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
	"git.casta.me/alberto/overmind/pkg/asciidoc"
)

const (
	headerPrefix    = "overmind-"
	headerID        = headerPrefix + "id"
	headerKind      = headerPrefix + "type"
	headerTitle     = headerPrefix + "title"
	headerTags      = headerPrefix + "tags"
	headerCreatedAt = headerPrefix + "created-at"
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
	headers := make(map[string]string)
	for _, attribute := range processed.Analysis.Header.Attributes.All() {
		if strings.HasPrefix(attribute.Name, headerPrefix) {
			headers[attribute.Name] = attribute.Value
		}
	}
	idValue, managed := headers[headerID]
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
	kind, err := domain.NewDocumentKind(headers[headerKind])
	if err != nil {
		return ports.IndexedDocument{}, false, fmt.Errorf("invalid %s: %w", headerKind, err)
	}
	title, err := domain.NewTitle(headers[headerTitle])
	if err != nil {
		return ports.IndexedDocument{}, false, fmt.Errorf("invalid %s: %w", headerTitle, err)
	}
	tags, err := tagsFromHeader(headers[headerTags])
	if err != nil {
		return ports.IndexedDocument{}, false, fmt.Errorf("invalid %s: %w", headerTags, err)
	}
	createdAt, err := time.Parse(time.RFC3339, headers[headerCreatedAt])
	if err != nil {
		return ports.IndexedDocument{}, false, fmt.Errorf("invalid %s: %w", headerCreatedAt, err)
	}

	attributes := make(map[string]string)
	for name, value := range headers {
		switch name {
		case headerID, headerKind, headerTitle, headerTags, headerCreatedAt:
			continue
		default:
			attributes[strings.TrimPrefix(name, headerPrefix)] = value
		}
	}

	return ports.IndexedDocument{
		ID:         documentID,
		Path:       blob.Path,
		Kind:       kind,
		Title:      title.String(),
		Tags:       tags.Strings(),
		CreatedAt:  createdAt,
		Attributes: attributes,
	}, true, nil
}

func tagsFromHeader(value string) (domain.Tags, error) {
	if strings.TrimSpace(value) == "" {
		return domain.Tags{}, nil
	}
	values := strings.Split(value, ",")
	tags := make([]domain.Tag, len(values))
	for index := range values {
		tag, err := domain.NewTag(strings.TrimSpace(values[index]))
		if err != nil {
			return domain.Tags{}, err
		}
		tags[index] = tag
	}
	return domain.NewTags(tags...)
}
