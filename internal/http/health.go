package http

import (
	stdhttp "net/http"

	httpresponse "git.casta.me/alberto/overmind/internal/http/response"
)

func handleHealth(response stdhttp.ResponseWriter, _ *stdhttp.Request) {
	httpresponse.JSON(response, stdhttp.StatusOK, struct {
		Status string `json:"status"`
	}{Status: "ok"})
}
