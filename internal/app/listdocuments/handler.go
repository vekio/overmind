// Package listdocuments implements the query for listing indexed documents.
package listdocuments

import (
	"context"
	"fmt"
	"strings"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

// ListDocumentsHandler lists summaries from the document read model.
type ListDocumentsHandler struct {
	index ports.DocumentIndexLister
}

// NewListDocumentsHandler creates the query handler.
func NewListDocumentsHandler(index ports.DocumentIndexLister) *ListDocumentsHandler {
	if index == nil {
		panic("list documents handler requires document index lister")
	}
	return &ListDocumentsHandler{index: index}
}

// Handle validates the filters and returns matching document summaries.
func (handler *ListDocumentsHandler) Handle(
	ctx context.Context,
	query ListDocumentsQuery,
) (ListDocumentsResult, error) {
	filter, err := listFilter(query)
	if err != nil {
		return ListDocumentsResult{}, fmt.Errorf("list documents: %w", err)
	}

	indexedDocuments, err := handler.index.List(ctx, filter)
	if err != nil {
		return ListDocumentsResult{}, fmt.Errorf("list documents: %w", err)
	}

	documents := make([]DocumentSummary, len(indexedDocuments))
	for index, document := range indexedDocuments {
		documents[index] = DocumentSummary{
			ID:    document.ID,
			Type:  document.Kind,
			Title: document.Title,
			Area:  document.Area,
			Tags:  append([]string(nil), document.Tags...),
		}
	}
	return ListDocumentsResult{Documents: documents}, nil
}

func listFilter(query ListDocumentsQuery) (ports.ListIndexedDocumentsFilter, error) {
	var kind domain.DocumentKind
	var err error
	if strings.TrimSpace(query.Type) != "" {
		kind, err = domain.NewDocumentKind(query.Type)
		if err != nil {
			return ports.ListIndexedDocumentsFilter{}, fmt.Errorf("validate type: %w", err)
		}
	}

	tags := make([]domain.Tag, len(query.Tags))
	for index, value := range query.Tags {
		tags[index], err = domain.NewTag(value)
		if err != nil {
			return ports.ListIndexedDocumentsFilter{}, fmt.Errorf("validate tag %d: %w", index+1, err)
		}
	}
	normalizedTags, err := domain.NewTags(tags...)
	if err != nil {
		return ports.ListIndexedDocumentsFilter{}, fmt.Errorf("validate tags: %w", err)
	}

	var area domain.Area
	if strings.TrimSpace(query.Area) != "" {
		area, err = domain.NewArea(query.Area)
		if err != nil {
			return ports.ListIndexedDocumentsFilter{}, fmt.Errorf("validate area: %w", err)
		}
	}

	return ports.ListIndexedDocumentsFilter{
		Kind: kind, Title: strings.TrimSpace(query.Title), Area: area, Tags: normalizedTags,
	}, nil
}
