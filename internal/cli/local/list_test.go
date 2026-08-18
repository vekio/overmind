package local

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/listdocuments"
	"git.casta.me/alberto/overmind/internal/config"
	"git.casta.me/alberto/overmind/internal/domain"
)

type listDocumentsHandlerStub struct {
	query  listdocuments.ListDocumentsQuery
	result listdocuments.ListDocumentsResult
	err    error
}

func (handler *listDocumentsHandlerStub) Handle(_ context.Context, query listdocuments.ListDocumentsQuery) (listdocuments.ListDocumentsResult, error) {
	handler.query = query
	return handler.result, handler.err
}

func TestListDocumentsWritesPipelineFriendlyTSV(t *testing.T) {
	configPath := writeLocalConfig(t)
	firstID, _ := domain.NewDocumentID("first-id")
	secondID, _ := domain.NewDocumentID("second-id")
	handler := &listDocumentsHandlerStub{result: listdocuments.ListDocumentsResult{Documents: []listdocuments.DocumentSummary{
		{ID: firstID, Path: "page/first.adoc", Type: domain.DocumentKindPage, Tags: []string{"go", "ddd"}},
		{ID: secondID, Path: "page/second.adoc", Type: domain.DocumentKindPage},
	}}}
	command := newTestCommand(t, func(config.CLIConfig) (Runtime, error) {
		return runtimeStub{application: &app.Application{
			Queries: app.Queries{ListDocuments: handler},
		}}, nil
	})
	var output bytes.Buffer
	command.Writer = &output

	err := command.Run(context.Background(), []string{
		"overmind", "--config", configPath,
		"ls", "--type", "page", "--tag", "Go", "--tag", "DDD",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if handler.query.Type != "page" || len(handler.query.Tags) != 2 ||
		handler.query.Tags[0] != "Go" || handler.query.Tags[1] != "DDD" {
		t.Fatalf("query = %+v", handler.query)
	}
	if output.String() != "page/first.adoc\tpage\tgo,ddd\npage/second.adoc\tpage\t\n" {
		t.Fatalf("output = %q", output.String())
	}
}

func TestListDocumentsPreservesUseCaseError(t *testing.T) {
	useCaseErr := errors.New("list indexed documents: database unavailable")
	handler := &listDocumentsHandlerStub{err: useCaseErr}
	state := newTestApplicationState(runtimeStub{application: &app.Application{
		Queries: app.Queries{ListDocuments: handler},
	}})

	err := newListCommand(state).Run(context.Background(), []string{"ls"})
	if !errors.Is(err, useCaseErr) || !strings.Contains(err.Error(), "database unavailable") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestListDocumentsRejectsArguments(t *testing.T) {
	err := newListCommand(&applicationState{}).Run(context.Background(), []string{"ls", "page"})
	if err == nil || !strings.Contains(err.Error(), "unexpected arguments") {
		t.Fatalf("Run() error = %v", err)
	}
}
