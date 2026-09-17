// Package bootstrap builds Overmind's dependency graph.
package bootstrap

import (
	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/app/listdocuments"
	"git.casta.me/alberto/overmind/internal/app/rebuildindex"
	"git.casta.me/alberto/overmind/internal/app/updatedocument"
	"git.casta.me/alberto/overmind/internal/config"
)

// New composes the application and the resources it owns.
func New(cfg config.Config) (Application, error) {
	deps, err := newDeps(cfg)
	if err != nil {
		return Application{}, err
	}

	return Application{
		Application: newApplication(deps),
		close:       deps.index.Close,
	}, nil
}

func newApplication(deps deps) app.Application {
	return app.Application{
		Commands: app.Commands{
			CreatePage: createpage.NewCreatePageHandler(
				deps.blobs, deps.renderer, deps.ids, deps.clock, deps.index,
			),
			RebuildIndex: rebuildindex.NewRebuildIndexHandler(deps.blobs, deps.index),
			UpdateDocument: updatedocument.NewUpdateDocumentHandler(
				deps.blobs, deps.blobs, deps.clock, deps.index,
			),
		},
		Queries: app.Queries{
			GetDocument:   getdocument.NewGetDocumentHandler(deps.blobs),
			ListDocuments: listdocuments.NewListDocumentsHandler(deps.index),
		},
	}
}
