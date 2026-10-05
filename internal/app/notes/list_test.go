package notes_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/vekio/overmind/internal/app"
	appnotes "github.com/vekio/overmind/internal/app/notes"
	"github.com/vekio/overmind/internal/ports"
)

type testNoteFinder struct {
	ports.Index
	calls  int
	filter ports.NoteFilter
	notes  []ports.NoteSummary
	err    error
}

func (finder *testNoteFinder) FindNotes(_ context.Context, filter ports.NoteFilter) ([]ports.NoteSummary, error) {
	finder.calls++
	finder.filter = filter
	return finder.notes, finder.err
}

func TestListNotesNormalizesFiltersAndAppliesDefaultLimit(t *testing.T) {
	finder := &testNoteFinder{notes: []ports.NoteSummary{{Type: "habit", Label: "Beber agua"}}}
	application := app.New(app.Dependencies{Index: finder})
	result, err := application.Queries.ListNotes.Handle(context.Background(), appnotes.ListQuery{Type: " HABIT ", Tag: " Sálud "})
	if err != nil {
		t.Fatal(err)
	}
	if finder.filter != (ports.NoteFilter{Type: "habit", Tag: "salud", Limit: 100}) || !reflect.DeepEqual(result.Notes, finder.notes) {
		t.Fatalf("filter=%+v result=%+v", finder.filter, result)
	}
}

func TestListNotesRejectsInvalidInputBeforeQuerying(t *testing.T) {
	for _, query := range []appnotes.ListQuery{
		{Type: "unknown"}, {Tag: "!!!"}, {Limit: -1}, {Limit: 1001}, {Offset: -1},
	} {
		finder := &testNoteFinder{}
		application := app.New(app.Dependencies{Index: finder})
		if _, err := application.Queries.ListNotes.Handle(context.Background(), query); err == nil || finder.calls != 0 {
			t.Fatalf("query=%+v err=%v calls=%d", query, err, finder.calls)
		}
	}
}

func TestListNotesPropagatesCancellationAndFinderErrors(t *testing.T) {
	failed := errors.New("index unavailable")
	finder := &testNoteFinder{err: failed}
	application := app.New(app.Dependencies{Index: finder})
	if _, err := application.Queries.ListNotes.Handle(context.Background(), appnotes.ListQuery{Limit: 25, Offset: 10}); !errors.Is(err, failed) {
		t.Fatal(err)
	}
	if finder.filter.Limit != 25 || finder.filter.Offset != 10 {
		t.Fatal(finder.filter)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := application.Queries.ListNotes.Handle(ctx, appnotes.ListQuery{}); !errors.Is(err, context.Canceled) || finder.calls != 1 {
		t.Fatalf("err=%v calls=%d", err, finder.calls)
	}
}
