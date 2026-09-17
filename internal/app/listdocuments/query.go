package listdocuments

import "git.casta.me/alberto/overmind/internal/domain"

// ListDocumentsQuery contains the optional filters accepted by the document
// listing use case. A document must contain every requested tag.
type ListDocumentsQuery struct {
	Type  string
	Title string
	Area  string
	Tags  []string
}

// DocumentSummary contains the fields needed to identify a listed document.
type DocumentSummary struct {
	ID    domain.DocumentID
	Type  domain.DocumentKind
	Title string
	Area  string
	Tags  []string
}

// ListDocumentsResult contains the matching documents in index order.
type ListDocumentsResult struct {
	Documents []DocumentSummary
}
