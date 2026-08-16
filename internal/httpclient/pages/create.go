// Package pages implements remote page use cases over HTTP.
package pages

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/httpclient"
)

const maxResponseSize = 1 << 20

var _ app.CreatePageHandler = (*CreateHandler)(nil)

// CreateHandler creates pages through a remote Overmind API.
type CreateHandler struct {
	client *httpclient.Client
}

// NewCreateHandler creates a remote create-page handler.
func NewCreateHandler(client *httpclient.Client) *CreateHandler {
	if client == nil {
		panic("remote create page handler requires HTTP client")
	}
	return &CreateHandler{client: client}
}

type createRequest struct {
	Title string   `json:"title"`
	Area  string   `json:"area"`
	Tags  []string `json:"tags,omitempty"`
}

type createResponse struct {
	ID string `json:"id"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Handle sends the create-page command to the remote API.
func (handler *CreateHandler) Handle(ctx context.Context, command createpage.CreatePageCommand) (createpage.CreatePageResult, error) {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(createRequest{Title: command.Title, Area: command.Area, Tags: command.Tags}); err != nil {
		return createpage.CreatePageResult{}, fmt.Errorf("encode create page request: %w", err)
	}

	request, err := handler.client.NewRequest(ctx, stdhttp.MethodPost, "pages", &body)
	if err != nil {
		return createpage.CreatePageResult{}, fmt.Errorf("create page: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := handler.client.Do(request)
	if err != nil {
		return createpage.CreatePageResult{}, fmt.Errorf("create page: %w", err)
	}
	defer response.Body.Close()

	responseBody := io.LimitReader(response.Body, maxResponseSize)
	if response.StatusCode != stdhttp.StatusCreated {
		var remoteError errorResponse
		if err := json.NewDecoder(responseBody).Decode(&remoteError); err != nil || remoteError.Error == "" {
			remoteError.Error = stdhttp.StatusText(response.StatusCode)
		}
		return createpage.CreatePageResult{}, &httpclient.ResponseError{
			StatusCode: response.StatusCode,
			Message:    remoteError.Error,
		}
	}

	var output createResponse
	if err := json.NewDecoder(responseBody).Decode(&output); err != nil {
		return createpage.CreatePageResult{}, fmt.Errorf("decode create page response: %w", err)
	}
	documentID, err := domain.NewDocumentID(output.ID)
	if err != nil {
		return createpage.CreatePageResult{}, fmt.Errorf("decode create page id: %w", err)
	}

	return createpage.CreatePageResult{ID: documentID}, nil
}
