// Package documents implements the HTTP adapter for document queries.
package documents

import (
	stdhttp "net/http"

	"git.casta.me/alberto/overmind/internal/app"
)

// Register adds document routes to mux.
func Register(mux *stdhttp.ServeMux, getDocument app.GetDocumentHandler) {
	if mux == nil {
		panic("document routes require HTTP mux")
	}
	mux.HandleFunc("GET /documents/{id}", handleGet(getDocument))
}
