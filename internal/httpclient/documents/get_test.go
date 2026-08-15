package documents

import (
	"context"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/httpclient"
)

func TestGetHandlerCallsRemoteAPI(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, request *stdhttp.Request) {
		if request.Method != stdhttp.MethodGet || request.URL.Path != "/api/documents/page-id" {
			t.Fatalf("request = %s %s", request.Method, request.URL.Path)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"id":"page-id","path":"page.adoc","content":"= Page\n","attributes":{"overmind-type":"page"}}`))
	}))
	defer server.Close()

	client, err := httpclient.New(server.URL+"/api", server.Client())
	if err != nil {
		t.Fatalf("httpclient.New() error = %v", err)
	}
	result, err := NewGetHandler(client).Handle(context.Background(), getdocument.GetDocumentQuery{ID: "page-id"})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.ID.String() != "page-id" || result.Path != "page.adoc" || string(result.Content) != "= Page\n" || result.Attributes[domain.AttributeType] != "page" {
		t.Fatalf("Handle() = %+v", result)
	}
}

func TestGetHandlerReturnsRemoteError(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, _ *stdhttp.Request) {
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(stdhttp.StatusNotFound)
		_, _ = response.Write([]byte(`{"error":"document not found"}`))
	}))
	defer server.Close()

	client, err := httpclient.New(server.URL, server.Client())
	if err != nil {
		t.Fatalf("httpclient.New() error = %v", err)
	}
	_, err = NewGetHandler(client).Handle(context.Background(), getdocument.GetDocumentQuery{ID: "missing"})
	var responseError *httpclient.ResponseError
	if !errors.As(err, &responseError) {
		t.Fatalf("Handle() error = %v, want ResponseError", err)
	}
	if responseError.StatusCode != stdhttp.StatusNotFound || responseError.Message != "document not found" {
		t.Fatalf("ResponseError = %+v", responseError)
	}
}
