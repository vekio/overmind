// Package getdocument implements retrieving a document from its source store.
package getdocument

import "git.casta.me/alberto/overmind/internal/domain"

// GetDocumentQuery identifies a document by its stable ID.
type GetDocumentQuery struct {
	ID domain.DocumentID
}

// GetDocumentResult contains the raw source document.
type GetDocumentResult struct {
	ID       domain.DocumentID
	Content  []byte
	Revision string
}
