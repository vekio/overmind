package index_test

import (
	"context"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/shared"
	noteindex "github.com/vekio/overmind/internal/infra/index"
	"github.com/vekio/overmind/internal/ports"
)

func TestFindNotesStableOrderTagsAndPagination(t *testing.T) {
	ctx := context.Background()
	index, err := noteindex.New(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = index.Close() })
	title, _ := shared.NewTitle("Agua")
	unit, _ := habits.NewUnit("litros")
	goal, _ := habits.NewGoal(2, unit, calendar.Day)
	first, _ := shared.NewTag("Salud")
	second, _ := shared.NewTag("Diario")
	tags, _ := shared.NewTags(first, second)
	// Same timestamps, reverse insertion: ID must break the ordering tie.
	ids := []uuid.UUID{
		uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		uuid.MustParse("11111111-1111-4111-8111-111111111111"),
	}
	for _, id := range ids {
		habit, err := habits.NewHabit(id, title, goal, tags, fixtureMetadata())
		if err != nil {
			t.Fatal(err)
		}
		if err := index.UpsertHabit(ctx, habit, "/notes/"+id.String()+".adoc"); err != nil {
			t.Fatal(err)
		}
	}
	for offset := 0; offset < 3; offset++ {
		notes, err := index.FindNotes(ctx, ports.NoteFilter{Type: "habit", Tag: "salud", Limit: 1, Offset: offset})
		if err != nil {
			t.Fatal(err)
		}
		if len(notes) != 1 || notes[0].ID != ids[2-offset] || notes[0].Label != "Agua" || len(notes[0].Tags) != 2 || notes[0].Tags[0] != "salud" || notes[0].Tags[1] != "diario" {
			t.Fatalf("page %d: %+v", offset, notes)
		}
	}
	notes, err := index.FindNotes(ctx, ports.NoteFilter{Tag: "absent", Limit: 100})
	if err != nil || len(notes) != 0 {
		t.Fatalf("empty result=%+v err=%v", notes, err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := index.FindNotes(canceled, ports.NoteFilter{Limit: 100}); err == nil {
		t.Fatal("canceled query accepted")
	}
}
