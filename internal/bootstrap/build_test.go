package bootstrap

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/vekio/overmind/internal/app"
	appconfig "github.com/vekio/overmind/internal/config"
	"github.com/vekio/overmind/internal/infra/index/sqlitedb"
)

func TestBuildApplicationPersistsHabitAcrossRestarts(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	settings := appconfig.Settings{Mode: appconfig.ModeLocal, VaultPath: root}
	application, closers, err := buildApplication(ctx, settings)
	if err != nil {
		t.Fatal(err)
	}
	created, err := application.Commands.CreateHabit.Handle(ctx, app.CreateHabitCommand{Title: "Beber agua", Amount: 2, Unit: "litros", Period: "day"})
	if err != nil {
		for _, closer := range closers {
			_ = closer.Close()
		}
		t.Fatal(err)
	}
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			t.Fatal(err)
		}
	}
	id := created.Habit.ID()
	notePath := filepath.Join(root, "notes", id.String()+".adoc")
	if _, err := os.Stat(notePath); err != nil {
		t.Fatal(err)
	}
	_, resources, err := buildApplication(ctx, settings)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, closer := range resources {
			if err := closer.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	if _, err := os.Stat(notePath); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", filepath.Join(root, "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	row, err := sqlitedb.New(db).HabitByID(ctx, id.String())
	if err != nil || row.Title != "Beber agua" || row.Type != "habit" {
		t.Fatalf("reopened habit=%+v err=%v", row, err)
	}
}
