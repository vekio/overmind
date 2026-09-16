package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/app/updatedocument"
)

type updateDocumentHandlerStub struct {
	command updatedocument.UpdateDocumentCommand
	result  updatedocument.UpdateDocumentResult
	err     error
	calls   int
}

func (handler *updateDocumentHandlerStub) Handle(
	_ context.Context,
	command updatedocument.UpdateDocumentCommand,
) (updatedocument.UpdateDocumentResult, error) {
	handler.command = command
	handler.calls++
	return handler.result, handler.err
}

func TestEditReadsPathFromPipelineAndUpdatesDocument(t *testing.T) {
	editorPath := writeEditorScript(t, `printf '\nEdited\n' >> "$1"`)
	t.Setenv("VISUAL", editorPath)
	t.Setenv("EDITOR", "missing-editor")

	getHandler := &getDocumentHandlerStub{result: getdocument.GetDocumentResult{
		Path: "page/knowledge/page.adoc", Content: []byte("= Page\n"), Revision: "revision-1",
	}}
	updateHandler := &updateDocumentHandlerStub{}
	state := newTestApplicationState(runtimeStub{application: &app.Application{
		Commands: app.Commands{UpdateDocument: updateHandler},
		Queries:  app.Queries{GetDocument: getHandler},
	}})
	command := newEditCommand(state)
	command.Reader = strings.NewReader("page/knowledge/page.adoc\tpage\tgo,ddd\n")

	if err := command.Run(context.Background(), []string{"edit"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if getHandler.query.Path != "page/knowledge/page.adoc" || updateHandler.calls != 1 ||
		updateHandler.command.Path != "page/knowledge/page.adoc" ||
		updateHandler.command.ExpectedRevision != "revision-1" ||
		!strings.Contains(string(updateHandler.command.Content), "Edited") {
		t.Fatalf("get query = %+v, update command = %+v, calls = %d", getHandler.query, updateHandler.command, updateHandler.calls)
	}
}

func TestEditDoesNotUpdateUnchangedDocument(t *testing.T) {
	editorPath := writeEditorScript(t, "exit 0")
	t.Setenv("VISUAL", editorPath)

	getHandler := &getDocumentHandlerStub{result: getdocument.GetDocumentResult{
		Path: "page/page.adoc", Content: []byte("= Page\n"), Revision: "revision-1",
	}}
	updateHandler := &updateDocumentHandlerStub{}
	state := newTestApplicationState(runtimeStub{application: &app.Application{
		Commands: app.Commands{UpdateDocument: updateHandler},
		Queries:  app.Queries{GetDocument: getHandler},
	}})

	if err := newEditCommand(state).Run(context.Background(), []string{"edit", "page/page.adoc"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if updateHandler.calls != 0 {
		t.Fatalf("UpdateDocument.Handle() calls = %d, want 0", updateHandler.calls)
	}
}

func TestEditPreservesUpdateFailure(t *testing.T) {
	editorPath := writeEditorScript(t, `printf 'changed' > "$1"`)
	t.Setenv("VISUAL", editorPath)
	updateErr := errors.New("document changed since it was read")
	state := newTestApplicationState(runtimeStub{application: &app.Application{
		Commands: app.Commands{UpdateDocument: &updateDocumentHandlerStub{err: updateErr}},
		Queries: app.Queries{GetDocument: &getDocumentHandlerStub{result: getdocument.GetDocumentResult{
			Path: "page/page.adoc", Content: []byte("old"), Revision: "revision-1",
		}}},
	}})

	err := newEditCommand(state).Run(context.Background(), []string{"edit", "page/page.adoc"})
	if !errors.Is(err, updateErr) {
		t.Fatalf("Run() error = %v, want %v", err, updateErr)
	}
}

func TestResolveEditorUsesVisualThenEditorThenFallback(t *testing.T) {
	visual := writeEditorScript(t, "exit 0")
	editor := writeEditorScript(t, "exit 0")
	t.Setenv("VISUAL", visual+" --wait")
	t.Setenv("EDITOR", editor)

	resolved, err := resolveEditor()
	if err != nil {
		t.Fatalf("resolveEditor() error = %v", err)
	}
	if resolved.executable != visual || len(resolved.arguments) != 1 || resolved.arguments[0] != "--wait" {
		t.Fatalf("resolveEditor() = %+v", resolved)
	}

	t.Setenv("VISUAL", "missing-visual-editor")
	resolved, err = resolveEditor()
	if err != nil || resolved.executable != editor {
		t.Fatalf("resolveEditor() = (%+v, %v), want EDITOR", resolved, err)
	}

	t.Setenv("EDITOR", "missing-editor")
	resolved, err = resolveEditor()
	if err != nil || filepath.Base(resolved.executable) != "vi" {
		t.Fatalf("resolveEditor() = (%+v, %v), want vi fallback", resolved, err)
	}
}

func writeEditorScript(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "editor")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o700); err != nil {
		t.Fatalf("WriteFile(editor) error = %v", err)
	}
	return path
}
