// Package app exposes the application's use-case handlers to input adapters.
package app

import (
	"context"

	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/app/rebuildindex"
)

// CreatePageHandler is the create-page command capability exposed to input
// adapters.
type CreatePageHandler interface {
	Handle(context.Context, createpage.CreatePageCommand) (createpage.CreatePageResult, error)
}

// RebuildIndexHandler is the rebuild-index command capability exposed to
// input adapters.
type RebuildIndexHandler interface {
	Handle(context.Context, rebuildindex.RebuildIndexCommand) (rebuildindex.RebuildIndexResult, error)
}

// GetDocumentHandler is the get-document query capability exposed to input
// adapters.
type GetDocumentHandler interface {
	Handle(context.Context, getdocument.GetDocumentQuery) (getdocument.GetDocumentResult, error)
}

// Application groups the use-case handlers exposed by Overmind. It contains
// no transport or infrastructure logic; CLI and HTTP may consume the same
// instance.
type Application struct {
	Commands Commands
	Queries  Queries
}

// Commands contains the application's state-changing handlers.
type Commands struct {
	CreatePage   CreatePageHandler
	RebuildIndex RebuildIndexHandler
}

// Queries contains the application's read-only handlers.
type Queries struct {
	GetDocument GetDocumentHandler
}
