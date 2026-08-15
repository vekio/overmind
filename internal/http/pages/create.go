package pages

import (
	"encoding/json"
	"errors"
	"io"
	stdhttp "net/http"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/domain"
	httpresponse "git.casta.me/alberto/overmind/internal/http/response"
	"git.casta.me/alberto/overmind/internal/ports"
)

const maxCreateRequestSize = 1 << 20

type createRequest struct {
	Title string `json:"title"`
	Area  string `json:"area"`
}

type createResponse struct {
	ID string `json:"id"`
}

func handleCreate(handler app.CreatePageHandler) stdhttp.HandlerFunc {
	if handler == nil {
		panic("create page HTTP handler requires use case")
	}

	return func(response stdhttp.ResponseWriter, request *stdhttp.Request) {
		request.Body = stdhttp.MaxBytesReader(response, request.Body, maxCreateRequestSize)
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()

		var input createRequest
		if err := decoder.Decode(&input); err != nil {
			httpresponse.Error(response, stdhttp.StatusBadRequest, "invalid JSON request")
			return
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			httpresponse.Error(response, stdhttp.StatusBadRequest, "request must contain one JSON object")
			return
		}

		result, err := handler.Handle(request.Context(), createpage.CreatePageCommand{
			Title: input.Title,
			Area:  input.Area,
		})
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidTitle), errors.Is(err, domain.ErrInvalidArea):
				httpresponse.Error(response, stdhttp.StatusBadRequest, "invalid page")
			case errors.Is(err, ports.ErrBlobAlreadyExists):
				httpresponse.Error(response, stdhttp.StatusConflict, "page already exists")
			default:
				httpresponse.Error(response, stdhttp.StatusInternalServerError, "internal server error")
			}
			return
		}

		httpresponse.JSON(response, stdhttp.StatusCreated, createResponse{ID: result.ID.String()})
	}
}
