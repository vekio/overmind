package ports

import (
	"context"
	"errors"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
)

var ErrIndexedDocumentNotFound = errors.New("indexed document not found")

// IndexedDocument is the query-side representation derived from an Overmind
// AsciiDoc document.
type IndexedDocument struct {
	ID         domain.DocumentID
	Kind       domain.DocumentKind
	Title      string
	Area       string
	Tags       []string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Attributes map[string]string
}

// DocumentIndexWriter updates the document read model.
type DocumentIndexWriter interface {
	Upsert(context.Context, IndexedDocument) error
	ReplaceAll(context.Context, []IndexedDocument) error
}

// DocumentIndexReader queries the document read model.
type DocumentIndexReader interface {
	GetByID(context.Context, domain.DocumentID) (IndexedDocument, error)
}

// ListIndexedDocumentsFilter restricts documents returned from the read
// model. Every requested tag must be present.
type ListIndexedDocumentsFilter struct {
	Kind  domain.DocumentKind
	Title string
	Area  domain.Area
	Tags  domain.Tags
}

// IndexedDocumentSummary is the lightweight representation returned by
// document listings.
type IndexedDocumentSummary struct {
	ID    domain.DocumentID
	Kind  domain.DocumentKind
	Title string
	Area  string
	Tags  []string
}

// DocumentIndexLister lists documents from the read model.
type DocumentIndexLister interface {
	List(context.Context, ListIndexedDocumentsFilter) ([]IndexedDocumentSummary, error)
}

// DocumentIndex combines both sides for adapters implementing the complete
// read model. Use cases should depend on the narrower interface.
type DocumentIndex interface {
	DocumentIndexWriter
	DocumentIndexReader
	DocumentIndexLister
}
