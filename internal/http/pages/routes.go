// Package pages implements the HTTP adapter for page use cases.
package pages

import (
	"log/slog"
	stdhttp "net/http"

	"git.casta.me/alberto/overmind/internal/app"
)

// Register adds page routes to mux.
func Register(mux *stdhttp.ServeMux, createPage app.CreatePageHandler, logger *slog.Logger) {
	if mux == nil {
		panic("page routes require HTTP mux")
	}
	if logger == nil {
		panic("page routes require logger")
	}
	mux.HandleFunc("POST /pages", handleCreate(createPage, logger))
}
