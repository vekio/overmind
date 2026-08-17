package app

import (
	"context"

	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/app/listdocuments"
	"git.casta.me/alberto/overmind/internal/app/rebuildindex"
	"git.casta.me/alberto/overmind/internal/app/updatedocument"
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

// UpdateDocumentHandler is the raw document-update command capability exposed
// to input adapters.
type UpdateDocumentHandler interface {
	Handle(context.Context, updatedocument.UpdateDocumentCommand) (updatedocument.UpdateDocumentResult, error)
}

// GetDocumentHandler is the get-document query capability exposed to input
// adapters.
type GetDocumentHandler interface {
	Handle(context.Context, getdocument.GetDocumentQuery) (getdocument.GetDocumentResult, error)
}

// ListDocumentsHandler is the list-documents query capability exposed to
// input adapters.
type ListDocumentsHandler interface {
	Handle(context.Context, listdocuments.ListDocumentsQuery) (listdocuments.ListDocumentsResult, error)
}
