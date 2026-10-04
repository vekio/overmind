package app_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/ports"
)

type deletionStore struct {
	ports.NoteStore
	calls *[]string
	err   error
}

func (store *deletionStore) Delete(context.Context, uuid.UUID) error {
	*store.calls = append(*store.calls, "store")
	return store.err
}

type deletionIndex struct {
	calls *[]string
	err   error
}

func (index *deletionIndex) Delete(context.Context, uuid.UUID) error {
	*index.calls = append(*index.calls, "index")
	return index.err
}

func TestDeleteNoteRemovesSourceBeforeIndexAndSupportsRetry(t *testing.T) {
	var calls []string
	indexErr := errors.New("index offline")
	store := &deletionStore{calls: &calls}
	index := &deletionIndex{calls: &calls, err: indexErr}
	application := app.New(app.Dependencies{NoteStore: store, IndexDeleter: index})
	id := uuid.New()
	result, err := application.Commands.DeleteNote.Handle(context.Background(), app.DeleteNoteCommand{ID: id.String()})
	if !errors.Is(err, indexErr) || result.ID != uuid.Nil() || !strings.Contains(err.Error(), "document removed") || !reflect.DeepEqual(calls, []string{"store", "index"}) {
		t.Fatalf("result=%+v err=%v calls=%v", result, err, calls)
	}
	index.err = nil
	result, err = application.Commands.DeleteNote.Handle(context.Background(), app.DeleteNoteCommand{ID: " " + id.String() + " "})
	if err != nil || result.ID != id || !reflect.DeepEqual(calls, []string{"store", "index", "store", "index"}) {
		t.Fatalf("retry result=%+v err=%v calls=%v", result, err, calls)
	}
}

func TestDeleteNoteStoreFailureLeavesIndexUntouched(t *testing.T) {
	var calls []string
	storeErr := errors.New("cannot remove file")
	application := app.New(app.Dependencies{NoteStore: &deletionStore{calls: &calls, err: storeErr}, IndexDeleter: &deletionIndex{calls: &calls}})
	_, err := application.Commands.DeleteNote.Handle(context.Background(), app.DeleteNoteCommand{ID: uuid.New().String()})
	if !errors.Is(err, storeErr) || !reflect.DeepEqual(calls, []string{"store"}) {
		t.Fatalf("err=%v calls=%v", err, calls)
	}
}

func TestDeleteNoteRejectsInvalidIDsAndCancellationBeforeMutation(t *testing.T) {
	var calls []string
	application := app.New(app.Dependencies{NoteStore: &deletionStore{calls: &calls}, IndexDeleter: &deletionIndex{calls: &calls}})
	for _, id := range []string{"", "invalid", "../../file", uuid.Nil().String()} {
		if _, err := application.Commands.DeleteNote.Handle(context.Background(), app.DeleteNoteCommand{ID: id}); err == nil {
			t.Fatalf("accepted %q", id)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := application.Commands.DeleteNote.Handle(ctx, app.DeleteNoteCommand{ID: uuid.New().String()}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if len(calls) != 0 {
		t.Fatalf("invalid or canceled command mutated storage: %v", calls)
	}
}
