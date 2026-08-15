package http

import (
	stdhttp "net/http"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/http/documents"
	"git.casta.me/alberto/overmind/internal/http/pages"
)

func routes(application *app.Application) stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	documents.Register(mux, application.Queries.GetDocument)
	pages.Register(mux, application.Commands.CreatePage)
	return mux
}
