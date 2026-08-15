package http

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
)

type createPageHandlerStub struct{}

func (createPageHandlerStub) Handle(context.Context, createpage.CreatePageCommand) (createpage.CreatePageResult, error) {
	return createpage.CreatePageResult{}, nil
}

type getDocumentHandlerStub struct{}

func (getDocumentHandlerStub) Handle(context.Context, getdocument.GetDocumentQuery) (getdocument.GetDocumentResult, error) {
	return getdocument.GetDocumentResult{}, nil
}

func testApplication() *app.Application {
	return &app.Application{
		Commands: app.Commands{CreatePage: createPageHandlerStub{}},
		Queries:  app.Queries{GetDocument: getDocumentHandlerStub{}},
	}
}

func TestNewServerConfiguresStandardServer(t *testing.T) {
	application := testApplication()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server, err := NewServer("127.0.0.1:8080", application, logger)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server.server.Addr != "127.0.0.1:8080" || server.server.Handler == nil {
		t.Fatalf("server = %#v", server.server)
	}
}

func TestNewServerRejectsInvalidAddress(t *testing.T) {
	application := testApplication()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if _, err := NewServer("localhost", application, logger); err == nil {
		t.Fatal("NewServer() error = nil")
	}
}
