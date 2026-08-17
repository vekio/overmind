package pages

import (
	"context"
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/domain"
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
		_, _ = response.Write([]byte(`{"id":"page-id","path":"page/knowledge/remote-page.adoc"}`))
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
	if result.Path != "page/knowledge/remote-page.adoc" {
		t.Fatalf("Path = %q", result.Path)
	}
}

func TestCreateHandlerReturnsPageConflict(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, _ *stdhttp.Request) {
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(stdhttp.StatusConflict)
		_, _ = response.Write([]byte(`{"error":"page already exists","code":"page_already_exists","path":"page/page.adoc"}`))
	}))
	defer server.Close()

	client, err := httpclient.New(server.URL, server.Client())
	if err != nil {
		t.Fatalf("httpclient.New() error = %v", err)
	}
	_, err = NewCreateHandler(client).Handle(context.Background(), createpage.CreatePageCommand{Title: "Page"})
	conflict, ok := errors.AsType[*createpage.PageAlreadyExistsError](err)
	if !ok || conflict.Path != "page/page.adoc" {
		t.Fatalf("Handle() error = %#v", err)
	}
}

func TestCreateHandlerReturnsIncompleteCreation(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, _ *stdhttp.Request) {
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(stdhttp.StatusInternalServerError)
		_, _ = response.Write([]byte(`{"error":"page creation incomplete","code":"page_creation_incomplete","path":"page/page.adoc"}`))
	}))
	defer server.Close()

	client, err := httpclient.New(server.URL, server.Client())
	if err != nil {
		t.Fatalf("httpclient.New() error = %v", err)
	}
	_, err = NewCreateHandler(client).Handle(context.Background(), createpage.CreatePageCommand{Title: "Page"})
	incomplete, ok := errors.AsType[*createpage.PageCreationIncompleteError](err)
	if !ok || incomplete.Path != "page/page.adoc" {
		t.Fatalf("Handle() error = %#v", err)
	}
}

func TestCreateHandlerReturnsDomainValidationErrors(t *testing.T) {
	for code, cause := range map[string]error{
		"invalid_page_title": domain.ErrInvalidTitle,
		"invalid_page_area":  domain.ErrInvalidArea,
		"invalid_page_tag":   domain.ErrInvalidTag,
		"duplicate_page_tag": domain.ErrDuplicateTag,
	} {
		t.Run(code, func(t *testing.T) {
			server := httptest.NewServer(stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, _ *stdhttp.Request) {
				response.Header().Set("Content-Type", "application/json")
				response.WriteHeader(stdhttp.StatusBadRequest)
				_ = json.NewEncoder(response).Encode(errorResponse{Error: "invalid page", Code: code})
			}))
			defer server.Close()

			client, err := httpclient.New(server.URL, server.Client())
			if err != nil {
				t.Fatalf("httpclient.New() error = %v", err)
			}
			_, err = NewCreateHandler(client).Handle(context.Background(), createpage.CreatePageCommand{Title: "Page"})
			if !errors.Is(err, cause) {
				t.Fatalf("Handle() error = %v, want %v", err, cause)
			}
		})
	}
}

func TestCreateHandlerReturnsUnexpectedRemoteError(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, _ *stdhttp.Request) {
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(stdhttp.StatusInternalServerError)
		_, _ = response.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	client, err := httpclient.New(server.URL, server.Client())
	if err != nil {
		t.Fatalf("httpclient.New() error = %v", err)
	}
	_, err = NewCreateHandler(client).Handle(context.Background(), createpage.CreatePageCommand{Title: "Page"})
	responseError, ok := errors.AsType[*httpclient.ResponseError](err)
	if !ok {
		t.Fatalf("Handle() error = %v, want ResponseError", err)
	}
	if responseError.StatusCode != stdhttp.StatusInternalServerError || responseError.Message != "internal server error" {
		t.Fatalf("ResponseError = %+v", responseError)
	}
}
