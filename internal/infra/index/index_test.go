package index_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/ports"
	"path/filepath"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/shared"
	noteindex "github.com/vekio/overmind/internal/infra/index"
	"github.com/vekio/overmind/internal/infra/index/sqlitedb"
)

func TestIndexPersistsHabitProjection(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "index.db")
	index, err := noteindex.New(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = index.Close() })
	title, _ := shared.NewTitle("Agua")
	unit, _ := habits.NewUnit("litros")
	goal, _ := habits.NewGoal(2, unit, calendar.Day)
	tag, _ := shared.NewTag("Salud")
	tags, _ := shared.NewTags(tag)
	habit, err := habits.NewHabit(uuid.New(), title, goal, tags, fixtureMetadata())
	if err != nil {
		t.Fatal(err)
	}
	if err := index.UpsertHabit(ctx, habit, "/vault/"+habit.ID().String()+".adoc"); err != nil {
		t.Fatal(err)
	}
	renamed, _ := shared.NewTitle("Beber agua")
	if err := habit.Rename(renamed, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := habit.ReplaceTags(shared.Tags{}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := index.UpsertHabit(ctx, habit, "/vault/"+habit.ID().String()+".adoc"); err != nil {
		t.Fatal(err)
	}
	if err := index.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	row, err := sqlitedb.New(db).HabitByID(ctx, habit.ID().String())
	if err != nil {
		t.Fatal(err)
	}
	if row.Title != "Beber agua" || row.TargetAmount != 2 || row.Unit != "litros" || row.Period != "day" {
		t.Fatal("habit projection did not persist replacements")
	}
	tagsAfter, err := sqlitedb.New(db).TagsByNoteID(ctx, habit.ID().String())
	if err != nil || len(tagsAfter) != 0 {
		t.Fatalf("replaced tags = %v, %v", tagsAfter, err)
	}
	if row.CreatedAt != "2026-10-03T12:00:00.123456789Z" || row.UpdatedAt == row.CreatedAt || row.Type != "habit" || row.Path != "/vault/"+habit.ID().String()+".adoc" {
		t.Fatal("common metadata did not persist")
	}
}

func fixtureMetadata() shared.EntityMetadata {
	at := time.Date(2026, 10, 3, 12, 0, 0, 123456789, time.UTC)
	metadata, err := shared.NewEntityMetadata(at, at)
	if err != nil {
		panic(err)
	}
	return metadata
}

func TestSpecializationFailureRollsBackCommonNoteAndTags(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "index.db")
	index, err := noteindex.New(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	title, _ := shared.NewTitle("Agua")
	unit, _ := habits.NewUnit("litros")
	goal, _ := habits.NewGoal(0.5, unit, calendar.Day)
	first, _ := shared.NewTag("Salud")
	second, _ := shared.NewTag("Bienestar")
	tags, _ := shared.NewTags(first, second)
	habit, _ := habits.NewHabit(uuid.New(), title, goal, tags, fixtureMetadata())
	if err := index.UpsertHabit(ctx, habit, "notes/agua.adoc"); err != nil {
		t.Fatal(err)
	}
	queries := sqlitedb.New(db)
	before, err := queries.HabitByID(ctx, habit.ID().String())
	if err != nil {
		t.Fatal(err)
	}
	indexedTags, err := queries.TagsByNoteID(ctx, habit.ID().String())
	if err != nil || len(indexedTags) != 2 || indexedTags[0] != "salud" || indexedTags[1] != "bienestar" {
		t.Fatalf("ordered tags = %v, %v", indexedTags, err)
	}
	if _, err := db.ExecContext(ctx, `CREATE TRIGGER reject_habit_update BEFORE UPDATE ON habits BEGIN SELECT RAISE(ABORT, 'forced failure'); END`); err != nil {
		t.Fatal(err)
	}
	renamed, _ := shared.NewTitle("Hidratarse")
	if err := habit.Rename(renamed, habit.Metadata().UpdatedAt().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := index.UpsertHabit(ctx, habit, "notes/new.adoc"); err == nil {
		t.Fatal("forced specialization failure ignored")
	}
	after, err := queries.HabitByID(ctx, habit.ID().String())
	if err != nil || after != before {
		t.Fatalf("common note was partially updated: %+v, %v", after, err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		t.Fatal(err)
	}
	if err := queries.DeleteNote(ctx, habit.ID().String()); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"habits", "note_tags"} {
		var count int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("cascade left %s rows: %d, %v", table, count, err)
		}
	}
}

func TestDuplicateJournalDateRollsBackSecondNote(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "index.db")
	index, err := noteindex.New(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	date, _ := calendar.NewDate("2024-02-29")
	first, _ := journals.NewJournal(uuid.New(), date, shared.Tags{}, fixtureMetadata())
	second, _ := journals.NewJournal(uuid.New(), date, shared.Tags{}, fixtureMetadata())
	if err := index.UpsertJournal(ctx, first, "first.adoc"); err != nil {
		t.Fatal(err)
	}
	if err := index.UpsertJournal(ctx, second, "second.adoc"); !errors.Is(err, ports.ErrJournalAlreadyExists) {
		t.Fatalf("duplicate date=%v", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM notes").Scan(&count); err != nil || count != 1 {
		t.Fatalf("second note not rolled back: count=%d err=%v", count, err)
	}
	var storedID string
	if err := db.QueryRowContext(ctx, "SELECT note_id FROM journals WHERE date=?", date.String()).Scan(&storedID); err != nil || storedID != first.ID().String() {
		t.Fatalf("original journal changed: id=%s err=%v", storedID, err)
	}
}
