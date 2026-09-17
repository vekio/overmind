// Package updatedocument implements replacing an existing Overmind document.
package updatedocument

import "git.casta.me/alberto/overmind/internal/domain"

// UpdateDocumentCommand contains the edited source and the revision originally
// read by the caller.
type UpdateDocumentCommand struct {
	ID               domain.DocumentID
	Content          []byte
	ExpectedRevision string
}

// UpdateDocumentResult identifies the stored document version.
type UpdateDocumentResult struct {
	ID       domain.DocumentID
	Revision string
}
