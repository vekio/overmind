package http

import (
	stdhttp "net/http"
)

func routes() stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	return mux
}
