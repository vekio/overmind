package documents

import (
	"errors"
	stdhttp "net/http"
	"time"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/domain"
	httpresponse "git.casta.me/alberto/overmind/internal/http/response"
	"git.casta.me/alberto/overmind/internal/ports"
)

type getResponse struct {
	ID         string            `json:"id"`
	Path       string            `json:"path"`
	Kind       string            `json:"kind"`
	Title      string            `json:"title"`
	Tags       []string          `json:"tags"`
	CreatedAt  time.Time         `json:"createdAt"`
	Content    string            `json:"content"`
	Attributes map[string]string `json:"attributes"`
}

func handleGet(handler app.GetDocumentHandler) stdhttp.HandlerFunc {
	if handler == nil {
		panic("get document HTTP handler requires use case")
	}

	return func(response stdhttp.ResponseWriter, request *stdhttp.Request) {
		result, err := handler.Handle(request.Context(), getdocument.GetDocumentQuery{ID: request.PathValue("id")})
		if err != nil {
			switch {
			case errors.Is(err, ports.ErrIndexedDocumentNotFound), errors.Is(err, ports.ErrBlobNotFound):
				httpresponse.Error(response, stdhttp.StatusNotFound, "document not found")
			case errors.Is(err, domain.ErrInvalidDocumentID):
				httpresponse.Error(response, stdhttp.StatusBadRequest, "invalid document id")
			default:
				httpresponse.Error(response, stdhttp.StatusInternalServerError, "internal server error")
			}
			return
		}

		httpresponse.JSON(response, stdhttp.StatusOK, getResponse{
			ID:         result.ID.String(),
			Path:       result.Path,
			Kind:       result.Kind.String(),
			Title:      result.Title,
			Tags:       result.Tags,
			CreatedAt:  result.CreatedAt,
			Content:    string(result.Content),
			Attributes: result.Attributes,
		})
	}
}
