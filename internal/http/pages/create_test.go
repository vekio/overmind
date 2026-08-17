package pages

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/domain"
)

type createHandlerStub struct {
	command createpage.CreatePageCommand
	result  createpage.CreatePageResult
	err     error
}

func (handler *createHandlerStub) Handle(_ context.Context, command createpage.CreatePageCommand) (createpage.CreatePageResult, error) {
	handler.command = command
	return handler.result, handler.err
}

func pageRoutes(handler *createHandlerStub) stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	Register(mux, handler, slog.New(slog.NewTextHandler(io.Discard, nil)))
	return mux
}

func TestCreate(t *testing.T) {
	id, _ := domain.NewDocumentID("page-id")
	handler := &createHandlerStub{result: createpage.CreatePageResult{ID: id, Path: "page/knowledge/first-page.adoc"}}
	request := httptest.NewRequest(stdhttp.MethodPost, "/pages", strings.NewReader(`{"title":"First page","area":"Knowledge","tags":["Go","DDD"]}`))
	response := httptest.NewRecorder()

	pageRoutes(handler).ServeHTTP(response, request)

	if response.Code != stdhttp.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, stdhttp.StatusCreated, response.Body)
	}
	if got, want := response.Body.String(), "{\"id\":\"page-id\",\"path\":\"page/knowledge/first-page.adoc\"}\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
	if got, want := response.Header().Get("Location"), "/documents/page-id"; got != want {
		t.Fatalf("Location = %q, want %q", got, want)
	}
	if handler.command.Title != "First page" || handler.command.Area != "Knowledge" || len(handler.command.Tags) != 2 {
		t.Fatalf("command = %+v", handler.command)
	}
}

func TestCreateRejectsInvalidJSON(t *testing.T) {
	request := httptest.NewRequest(stdhttp.MethodPost, "/pages", strings.NewReader(`{"unknown":true}`))
	response := httptest.NewRecorder()

	pageRoutes(&createHandlerStub{}).ServeHTTP(response, request)

	if response.Code != stdhttp.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusBadRequest)
	}
}

func TestCreateMapsUseCaseErrors(t *testing.T) {
	for name, test := range map[string]struct {
		err    error
		status int
		body   string
	}{
		"invalid title": {err: domain.ErrInvalidTitle, status: stdhttp.StatusBadRequest, body: `{"error":"invalid page title","code":"invalid_page_title"}` + "\n"},
		"invalid area":  {err: domain.ErrInvalidArea, status: stdhttp.StatusBadRequest, body: `{"error":"invalid page area","code":"invalid_page_area"}` + "\n"},
		"invalid tag":   {err: domain.ErrInvalidTag, status: stdhttp.StatusBadRequest, body: `{"error":"invalid page tag","code":"invalid_page_tag"}` + "\n"},
		"duplicate tag": {err: domain.ErrDuplicateTag, status: stdhttp.StatusBadRequest, body: `{"error":"duplicate page tag","code":"duplicate_page_tag"}` + "\n"},
		"conflict":      {err: &createpage.PageAlreadyExistsError{Path: "page/page.adoc"}, status: stdhttp.StatusConflict, body: `{"error":"page already exists","code":"page_already_exists","path":"page/page.adoc"}` + "\n"},
		"incomplete":    {err: createpage.NewPageCreationIncompleteError("page/page.adoc"), status: stdhttp.StatusInternalServerError, body: `{"error":"page creation incomplete","code":"page_creation_incomplete","path":"page/page.adoc"}` + "\n"},
		"internal":      {err: errors.New("failure"), status: stdhttp.StatusInternalServerError, body: `{"error":"internal server error"}` + "\n"},
	} {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(stdhttp.MethodPost, "/pages", strings.NewReader(`{"title":"Page"}`))
			response := httptest.NewRecorder()

			pageRoutes(&createHandlerStub{err: test.err}).ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			if response.Body.String() != test.body {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.body)
			}
		})
	}
}

func TestCreateLogsInternalErrors(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	request := httptest.NewRequest(stdhttp.MethodPost, "/pages", strings.NewReader(`{"title":"Page"}`))
	response := httptest.NewRecorder()

	handleCreate(&createHandlerStub{err: errors.New("database unavailable")}, logger).ServeHTTP(response, request)

	if response.Code != stdhttp.StatusInternalServerError || !strings.Contains(logs.String(), `"msg":"create page failed"`) || !strings.Contains(logs.String(), "database unavailable") {
		t.Fatalf("status = %d, logs = %s", response.Code, logs.String())
	}
}
