package documents

import (
	"context"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

type getHandlerStub struct {
	query  getdocument.GetDocumentQuery
	result getdocument.GetDocumentResult
	err    error
}

func (handler *getHandlerStub) Handle(_ context.Context, query getdocument.GetDocumentQuery) (getdocument.GetDocumentResult, error) {
	handler.query = query
	return handler.result, handler.err
}

func documentRoutes(handler *getHandlerStub) stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	Register(mux, handler)
	return mux
}

func TestGet(t *testing.T) {
	id, _ := domain.NewDocumentID("page-id")
	handler := &getHandlerStub{result: getdocument.GetDocumentResult{
		ID:      id,
		Path:    "knowledge/page.adoc",
		Content: []byte("= Page\n"),
		Attributes: map[string]string{
			domain.AttributeType: "page",
			domain.AttributeArea: "knowledge",
		},
	}}
	request := httptest.NewRequest(stdhttp.MethodGet, "/documents/page-id", nil)
	response := httptest.NewRecorder()

	documentRoutes(handler).ServeHTTP(response, request)

	if response.Code != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, stdhttp.StatusOK, response.Body)
	}
	if handler.query.ID != "page-id" {
		t.Fatalf("query = %+v", handler.query)
	}
	want := "{\"id\":\"page-id\",\"path\":\"knowledge/page.adoc\",\"content\":\"= Page\\n\",\"attributes\":{\"overmind-area\":\"knowledge\",\"overmind-type\":\"page\"}}\n"
	if response.Body.String() != want {
		t.Fatalf("body = %q, want %q", response.Body.String(), want)
	}
}

func TestGetMapsErrors(t *testing.T) {
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"invalid id": {err: domain.ErrInvalidDocumentID, status: stdhttp.StatusBadRequest},
		"not found":  {err: ports.ErrIndexedDocumentNotFound, status: stdhttp.StatusNotFound},
		"internal":   {err: errors.New("failure"), status: stdhttp.StatusInternalServerError},
	} {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(stdhttp.MethodGet, "/documents/page-id", nil)
			response := httptest.NewRecorder()
			documentRoutes(&getHandlerStub{err: test.err}).ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
		})
	}
}
