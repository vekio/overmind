package index_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/shared"
	noteindex "github.com/vekio/overmind/internal/infra/index"
	"github.com/vekio/overmind/internal/ports"
)

func TestRebuildRollsBackCompletedUpsertsOnCancellation(t *testing.T) {
	ctx := context.Background()
	index, err := noteindex.New(ctx, filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = index.Close() })
	title, _ := shared.NewTitle("Original")
	unit, _ := habits.NewUnit("litros")
	goal, _ := habits.NewGoal(2, unit, calendar.Day)
	original, _ := habits.NewHabit(uuid.New(), title, goal, shared.Tags{}, fixtureMetadata())
	if err := index.UpsertHabit(ctx, original, "/notes/original.adoc"); err != nil {
		t.Fatal(err)
	}
	replacement, _ := habits.NewHabit(uuid.New(), title, goal, shared.Tags{}, fixtureMetadata())
	canceled, cancel := context.WithCancel(ctx)
	err = index.Rebuild(canceled, func(scoped ports.Index) error {
		if err := scoped.UpsertHabit(canceled, replacement, "/notes/replacement.adoc"); err != nil {
			return err
		}
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
	notes, err := index.FindNotes(ctx, ports.NoteFilter{Limit: 100})
	if err != nil || len(notes) != 1 || notes[0].ID != original.ID() {
		t.Fatalf("rollback notes=%+v err=%v", notes, err)
	}
}
