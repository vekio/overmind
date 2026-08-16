// Package getdocument implements the query for retrieving an indexed document.
package getdocument

import (
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
)

// GetDocumentQuery identifies the document to retrieve.
type GetDocumentQuery struct {
	ID string
}

// GetDocumentResult is the generic query-side document representation.
type GetDocumentResult struct {
	ID         domain.DocumentID
	Path       string
	Kind       domain.DocumentKind
	Title      string
	CreatedAt  time.Time
	Content    []byte
	Attributes map[string]string
}
