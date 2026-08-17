// Package getdocument implements retrieving a document from its source store.
package getdocument

// GetDocumentQuery identifies a document by its store-relative path.
type GetDocumentQuery struct {
	Path string
}

// GetDocumentResult contains the raw source document.
type GetDocumentResult struct {
	Path     string
	Content  []byte
	Revision string
}
