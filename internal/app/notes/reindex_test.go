package notes_test

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/app/notes"
	"github.com/vekio/overmind/internal/ports"
)

type reindexScanner struct {
	ports.NoteStore
	notes []ports.Note
	err   error
}

func (scanner reindexScanner) Scan(ctx context.Context, visit func(ports.Note) error) error {
	for _, note := range scanner.notes {
		if err := visit(note); err != nil {
			return err
		}
	}
	return scanner.err
}

type reindexProjector struct {
	calls int
	err   error
}

func (projector *reindexProjector) Project(context.Context, ports.Note, ports.Index) error {
	projector.calls++
	return projector.err
}

type reindexBackend struct {
	ports.Index
	projector   *reindexProjector
	calls       int
	populateErr error
}

func (backend *reindexBackend) Rebuild(ctx context.Context, populate func(ports.Index) error) error {
	backend.calls++
	backend.populateErr = populate(backend)
	return backend.populateErr
}

func TestReindexCountsOnlySuccessfullyRebuiltDocuments(t *testing.T) {
	scanner := reindexScanner{notes: []ports.Note{{ID: uuid.New(), Path: "first.adoc"}, {ID: uuid.New(), Path: "second.adoc"}}}
	projector := &reindexProjector{}
	backend := &reindexBackend{projector: projector}
	application := app.New(app.Dependencies{NoteStore: scanner, Index: backend})
	result, err := application.Commands.Reindex.Handle(context.Background(), notes.ReindexCommand{})
	if err != nil || result.Indexed != 2 || backend.calls != 1 || projector.calls != 2 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestReindexRejectsDuplicateIDsAndPropagatesScanErrors(t *testing.T) {
	id := uuid.New()
	failed := errors.New("read failed")
	for _, scanner := range []reindexScanner{
		{notes: []ports.Note{{ID: id, Path: "a.adoc"}, {ID: id, Path: "b.adoc"}}},
		{notes: []ports.Note{{ID: id, Path: "a.adoc"}}, err: failed},
	} {
		backend := &reindexBackend{projector: &reindexProjector{}}
		application := app.New(app.Dependencies{NoteStore: scanner, Index: backend})
		result, err := application.Commands.Reindex.Handle(context.Background(), notes.ReindexCommand{})
		if err == nil || result.Indexed != 0 || backend.populateErr == nil {
			t.Fatalf("result=%+v err=%v", result, err)
		}
		if scanner.err != nil && !errors.Is(err, failed) {
			t.Fatalf("lost scanner error: %v", err)
		}
	}
}

func TestReindexCanceledBeforeStartingDoesNotTouchIndex(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	backend := &reindexBackend{projector: &reindexProjector{}}
	application := app.New(app.Dependencies{NoteStore: reindexScanner{}, Index: backend})
	_, err := application.Commands.Reindex.Handle(ctx, notes.ReindexCommand{})
	if !errors.Is(err, context.Canceled) || backend.calls != 0 {
		t.Fatalf("err=%v calls=%d", err, backend.calls)
	}
}

func (backend *reindexBackend) Project(ctx context.Context, note ports.Note) error {
	return backend.projector.Project(ctx, note, backend)
}
