package listdocuments

import "git.casta.me/alberto/overmind/internal/domain"

// ListDocumentsQuery contains the optional filters accepted by the document
// listing use case. A document must contain every requested tag.
type ListDocumentsQuery struct {
	Type       string
	Tags       []string
	PathPrefix string
}

// DocumentSummary contains the fields needed to identify a listed document.
type DocumentSummary struct {
	ID   domain.DocumentID
	Path string
	Type domain.DocumentKind
	Tags []string
}

// ListDocumentsResult contains the matching documents in index order.
type ListDocumentsResult struct {
	Documents []DocumentSummary
}
