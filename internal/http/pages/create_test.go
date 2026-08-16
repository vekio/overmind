package pages

import (
	"context"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
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
	Register(mux, handler)
	return mux
}

func TestCreate(t *testing.T) {
	id, _ := domain.NewDocumentID("page-id")
	handler := &createHandlerStub{result: createpage.CreatePageResult{ID: id}}
	request := httptest.NewRequest(stdhttp.MethodPost, "/pages", strings.NewReader(`{"title":"First page","area":"Knowledge","tags":["Go","DDD"]}`))
	response := httptest.NewRecorder()

	pageRoutes(handler).ServeHTTP(response, request)

	if response.Code != stdhttp.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, stdhttp.StatusCreated, response.Body)
	}
	if got, want := response.Body.String(), "{\"id\":\"page-id\"}\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
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
	}{
		"invalid page":  {err: domain.ErrInvalidTitle, status: stdhttp.StatusBadRequest},
		"invalid tag":   {err: domain.ErrInvalidTag, status: stdhttp.StatusBadRequest},
		"duplicate tag": {err: domain.ErrDuplicateTag, status: stdhttp.StatusBadRequest},
		"conflict":      {err: ports.ErrBlobAlreadyExists, status: stdhttp.StatusConflict},
		"internal":      {err: errors.New("failure"), status: stdhttp.StatusInternalServerError},
	} {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(stdhttp.MethodPost, "/pages", strings.NewReader(`{"title":"Page"}`))
			response := httptest.NewRecorder()

			pageRoutes(&createHandlerStub{err: test.err}).ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
		})
	}
}
