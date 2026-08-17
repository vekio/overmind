package pages

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	stdhttp "net/http"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/domain"
	httpresponse "git.casta.me/alberto/overmind/internal/http/response"
)

const maxCreateRequestSize = 1 << 20

type createRequest struct {
	Title string   `json:"title"`
	Area  string   `json:"area"`
	Tags  []string `json:"tags,omitempty"`
}

type createResponse struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type createErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
	Path  string `json:"path,omitempty"`
}

func handleCreate(handler app.CreatePageHandler, logger *slog.Logger) stdhttp.HandlerFunc {
	if handler == nil {
		panic("create page HTTP handler requires use case")
	}
	if logger == nil {
		panic("create page HTTP handler requires logger")
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
			Tags:  input.Tags,
		})
		if err != nil {
			conflict, isConflict := errors.AsType[*createpage.PageAlreadyExistsError](err)
			incomplete, isIncomplete := errors.AsType[*createpage.PageCreationIncompleteError](err)
			switch {
			case isConflict:
				httpresponse.JSON(response, stdhttp.StatusConflict, createErrorResponse{
					Error: "page already exists",
					Code:  "page_already_exists",
					Path:  conflict.Path,
				})
			case isIncomplete:
				logger.ErrorContext(request.Context(), "page creation incomplete", "path", incomplete.Path, "error", err)
				httpresponse.JSON(response, stdhttp.StatusInternalServerError, createErrorResponse{
					Error: "page creation incomplete",
					Code:  "page_creation_incomplete",
					Path:  incomplete.Path,
				})
			case errors.Is(err, domain.ErrInvalidTitle):
				httpresponse.JSON(response, stdhttp.StatusBadRequest, createErrorResponse{
					Error: "invalid page title",
					Code:  "invalid_page_title",
				})
			case errors.Is(err, domain.ErrInvalidArea):
				httpresponse.JSON(response, stdhttp.StatusBadRequest, createErrorResponse{
					Error: "invalid page area",
					Code:  "invalid_page_area",
				})
			case errors.Is(err, domain.ErrInvalidTag):
				httpresponse.JSON(response, stdhttp.StatusBadRequest, createErrorResponse{
					Error: "invalid page tag",
					Code:  "invalid_page_tag",
				})
			case errors.Is(err, domain.ErrDuplicateTag):
				httpresponse.JSON(response, stdhttp.StatusBadRequest, createErrorResponse{
					Error: "duplicate page tag",
					Code:  "duplicate_page_tag",
				})
			default:
				logger.ErrorContext(request.Context(), "create page failed", "error", err)
				httpresponse.Error(response, stdhttp.StatusInternalServerError, "internal server error")
			}
			return
		}

		response.Header().Set("Location", "/documents/"+result.ID.String())
		httpresponse.JSON(response, stdhttp.StatusCreated, createResponse{
			ID:   result.ID.String(),
			Path: result.Path,
		})
	}
}
