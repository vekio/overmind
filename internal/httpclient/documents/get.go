// Package documents implements remote document queries over HTTP.
package documents

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"time"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/httpclient"
)

const maxResponseSize = 1 << 20

var _ app.GetDocumentHandler = (*GetHandler)(nil)

// GetHandler retrieves documents through a remote Overmind API.
type GetHandler struct {
	client *httpclient.Client
}

// NewGetHandler creates a remote get-document handler.
func NewGetHandler(client *httpclient.Client) *GetHandler {
	if client == nil {
		panic("remote get document handler requires HTTP client")
	}
	return &GetHandler{client: client}
}

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

type errorResponse struct {
	Error string `json:"error"`
}

// Handle sends the get-document query to the remote API.
func (handler *GetHandler) Handle(ctx context.Context, query getdocument.GetDocumentQuery) (getdocument.GetDocumentResult, error) {
	request, err := handler.client.NewRequest(ctx, stdhttp.MethodGet, "documents/"+url.PathEscape(query.ID), nil)
	if err != nil {
		return getdocument.GetDocumentResult{}, fmt.Errorf("get document: %w", err)
	}
	request.Header.Set("Accept", "application/json")

	response, err := handler.client.Do(request)
	if err != nil {
		return getdocument.GetDocumentResult{}, fmt.Errorf("get document: %w", err)
	}
	defer response.Body.Close()
	responseBody := io.LimitReader(response.Body, maxResponseSize)
	if response.StatusCode != stdhttp.StatusOK {
		var remoteError errorResponse
		if err := json.NewDecoder(responseBody).Decode(&remoteError); err != nil || remoteError.Error == "" {
			remoteError.Error = stdhttp.StatusText(response.StatusCode)
		}
		return getdocument.GetDocumentResult{}, &httpclient.ResponseError{
			StatusCode: response.StatusCode,
			Message:    remoteError.Error,
		}
	}

	var output getResponse
	if err := json.NewDecoder(responseBody).Decode(&output); err != nil {
		return getdocument.GetDocumentResult{}, fmt.Errorf("decode get document response: %w", err)
	}
	documentID, err := domain.NewDocumentID(output.ID)
	if err != nil {
		return getdocument.GetDocumentResult{}, fmt.Errorf("decode get document id: %w", err)
	}
	kind, err := domain.NewDocumentKind(output.Kind)
	if err != nil {
		return getdocument.GetDocumentResult{}, fmt.Errorf("decode get document kind: %w", err)
	}
	return getdocument.GetDocumentResult{
		ID:         documentID,
		Path:       output.Path,
		Kind:       kind,
		Title:      output.Title,
		Tags:       output.Tags,
		CreatedAt:  output.CreatedAt,
		Content:    []byte(output.Content),
		Attributes: output.Attributes,
	}, nil
}
