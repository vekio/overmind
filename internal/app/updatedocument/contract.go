// Package updatedocument implements replacing an existing Overmind document.
package updatedocument

// UpdateDocumentCommand contains the edited source and the revision originally
// read by the caller.
type UpdateDocumentCommand struct {
	Path             string
	Content          []byte
	ExpectedRevision string
}

// UpdateDocumentResult identifies the stored document version.
type UpdateDocumentResult struct {
	Path     string
	Revision string
}
