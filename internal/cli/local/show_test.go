package local

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/app/listdocuments"
	"git.casta.me/alberto/overmind/internal/config"
	"git.casta.me/alberto/overmind/internal/domain"
)

type getDocumentHandlerStub struct {
	query  getdocument.GetDocumentQuery
	result getdocument.GetDocumentResult
}

func (handler *getDocumentHandlerStub) Handle(_ context.Context, query getdocument.GetDocumentQuery) (getdocument.GetDocumentResult, error) {
	handler.query = query
	return handler.result, nil
}

type showListDocumentsHandlerStub struct {
	query  listdocuments.ListDocumentsQuery
	result listdocuments.ListDocumentsResult
}

func (handler *showListDocumentsHandlerStub) Handle(_ context.Context, query listdocuments.ListDocumentsQuery) (listdocuments.ListDocumentsResult, error) {
	handler.query = query
	return handler.result, nil
}

func TestShowWritesRawDocumentByPath(t *testing.T) {
	configPath := writeLocalConfig(t)
	handler := &getDocumentHandlerStub{result: getdocument.GetDocumentResult{
		Path: "page/knowledge/page.adoc", Content: []byte("= Page\n"),
	}}
	command := newTestCommand(t, func(config.CLIConfig) (Runtime, error) {
		return runtimeStub{application: &app.Application{
			Queries: app.Queries{GetDocument: handler},
		}}, nil
	})
	var output bytes.Buffer
	command.Writer = &output

	if err := command.Run(context.Background(), []string{"overmind", "--config", configPath, "show", "page/knowledge/page.adoc"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if handler.query.Path != "page/knowledge/page.adoc" || output.String() != "= Page\n" {
		t.Fatalf("query = %+v, output = %q", handler.query, output.String())
	}
}

func TestShowReadsPathFromPipeline(t *testing.T) {
	handler := &getDocumentHandlerStub{result: getdocument.GetDocumentResult{
		Path: "page/knowledge/page.adoc", Content: []byte("= Page\n"),
	}}
	state := newTestApplicationState(runtimeStub{application: &app.Application{
		Queries: app.Queries{GetDocument: handler},
	}})
	command := newShowCommand(state)
	command.Reader = strings.NewReader("page/knowledge/page.adoc\tpage\tgo,ddd\n")
	var output bytes.Buffer
	command.Writer = &output

	if err := command.Run(context.Background(), []string{"show"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if handler.query.Path != "page/knowledge/page.adoc" || output.String() != "= Page\n" {
		t.Fatalf("query = %+v, output = %q", handler.query, output.String())
	}
}

func TestShowRejectsMultipleDocumentsFromPipeline(t *testing.T) {
	command := newShowCommand(&applicationState{})
	command.Reader = strings.NewReader("page/first.adoc\npage/second.adoc\n")
	err := command.Run(context.Background(), []string{"show"})
	if err == nil || !strings.Contains(err.Error(), "exactly one document") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestShowCompletesIndexedPathPrefix(t *testing.T) {
	configPath := writeLocalConfig(t)
	id, _ := domain.NewDocumentID("page-id")
	handler := &showListDocumentsHandlerStub{result: listdocuments.ListDocumentsResult{Documents: []listdocuments.DocumentSummary{{
		ID: id, Path: "page/development/go.adoc", Type: domain.DocumentKindPage,
	}}}}
	command := newTestCommand(t, func(config.CLIConfig) (Runtime, error) {
		return runtimeStub{application: &app.Application{
			Queries: app.Queries{ListDocuments: handler},
		}}, nil
	})
	var output bytes.Buffer
	command.Writer = &output

	err := command.Run(context.Background(), []string{
		"overmind", "--config", configPath,
		"show", "page/dev", "--generate-shell-completion",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if handler.query.PathPrefix != "page/dev" || output.String() != "page/development/go.adoc\n" {
		t.Fatalf("query = %+v, output = %q", handler.query, output.String())
	}
}

func TestShowRejectsArgumentsAfterPath(t *testing.T) {
	err := newShowCommand(&applicationState{}).Run(
		context.Background(),
		[]string{"show", "page/page.adoc", "extra"},
	)
	if err == nil || !strings.Contains(err.Error(), "unexpected arguments") {
		t.Fatalf("Run() error = %v, want unexpected arguments error", err)
	}
}
