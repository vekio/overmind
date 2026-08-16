package pages

import (
	"context"
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/httpclient"
)

func TestCreateHandlerCallsRemoteAPI(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, request *stdhttp.Request) {
		if request.Method != stdhttp.MethodPost || request.URL.Path != "/api/pages" {
			t.Fatalf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("Content-Type = %q", request.Header.Get("Content-Type"))
		}
		var input createRequest
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if input.Title != "Remote page" || input.Area != "Knowledge" || len(input.Tags) != 2 || input.Tags[0] != "Go" || input.Tags[1] != "DDD" {
			t.Fatalf("request body = %+v", input)
		}
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(stdhttp.StatusCreated)
		_, _ = response.Write([]byte(`{"id":"page-id"}`))
	}))
	defer server.Close()

	client, err := httpclient.New(server.URL+"/api", server.Client())
	if err != nil {
		t.Fatalf("httpclient.New() error = %v", err)
	}
	result, err := NewCreateHandler(client).Handle(context.Background(), createpage.CreatePageCommand{
		Title: "Remote page",
		Area:  "Knowledge",
		Tags:  []string{"Go", "DDD"},
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.ID.String() != "page-id" {
		t.Fatalf("ID = %q, want %q", result.ID, "page-id")
	}
}

func TestCreateHandlerReturnsRemoteError(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, _ *stdhttp.Request) {
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(stdhttp.StatusConflict)
		_, _ = response.Write([]byte(`{"error":"page already exists"}`))
	}))
	defer server.Close()

	client, err := httpclient.New(server.URL, server.Client())
	if err != nil {
		t.Fatalf("httpclient.New() error = %v", err)
	}
	_, err = NewCreateHandler(client).Handle(context.Background(), createpage.CreatePageCommand{Title: "Page"})
	var responseError *httpclient.ResponseError
	if !errors.As(err, &responseError) {
		t.Fatalf("Handle() error = %v, want ResponseError", err)
	}
	if responseError.StatusCode != stdhttp.StatusConflict || responseError.Message != "page already exists" {
		t.Fatalf("ResponseError = %+v", responseError)
	}
}
