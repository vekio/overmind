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
	Path       string
	Kind       domain.DocumentKind
	Title      string
	CreatedAt  time.Time
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

// DocumentIndex combines both sides for adapters implementing the complete
// read model. Use cases should depend on the narrower interface.
type DocumentIndex interface {
	DocumentIndexWriter
	DocumentIndexReader
}
