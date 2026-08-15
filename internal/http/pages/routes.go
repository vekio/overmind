// Package pages implements the HTTP adapter for page use cases.
package pages

import (
	stdhttp "net/http"

	"git.casta.me/alberto/overmind/internal/app"
)

// Register adds page routes to mux.
func Register(mux *stdhttp.ServeMux, createPage app.CreatePageHandler) {
	if mux == nil {
		panic("page routes require HTTP mux")
	}
	mux.HandleFunc("POST /pages", handleCreate(createPage))
}
