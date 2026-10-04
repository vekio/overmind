// Package index implements the note index using the current note projections.
package index

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/infra/index/sqlitedb"
	"github.com/vekio/overmind/internal/ports"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

//go:embed schema.sql
var schema string

var _ ports.Index = (*Index)(nil)

type Index struct {
	db          *sql.DB
	queries     *sqlitedb.Queries
	transaction *sql.Tx
}

func New(ctx context.Context, path string) (*Index, error) {
	if path == "" {
		return nil, fmt.Errorf("index path is required")
	}
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	for _, statement := range []string{"PRAGMA foreign_keys = ON", "PRAGMA busy_timeout = 5000", "PRAGMA journal_mode = WAL", schema} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			db.Close()
			return nil, fmt.Errorf("initialize note index: %w", err)
		}
	}
	return &Index{db: db, queries: sqlitedb.New(db)}, nil
}

func (index *Index) UpsertHabit(ctx context.Context, habit *habits.Habit, path string) error {
	if habit == nil || habit.ID() == uuid.Nil() {
		return fmt.Errorf("initialized habit is required")
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("note path is required")
	}
	transaction, err := index.beginWrite(ctx)
	if err != nil {
		return fmt.Errorf("begin habit index transaction: %w", err)
	}
	defer transaction.Rollback()
	queries := index.queries.WithTx(transaction.Tx)
	timestamp := func(at time.Time) string { return at.UTC().Format("2006-01-02T15:04:05.000000000Z") }
	if err := queries.UpsertNote(ctx, sqlitedb.UpsertNoteParams{
		ID: habit.ID().String(), Type: ports.NoteKindHabit.String(), Path: path,
		CreatedAt: timestamp(habit.Metadata().CreatedAt()), UpdatedAt: timestamp(habit.Metadata().UpdatedAt()),
	}); err != nil {
		return fmt.Errorf("upsert note: %w", err)
	}
	if err := queries.UpsertHabit(ctx, sqlitedb.UpsertHabitParams{NoteID: habit.ID().String(), Title: habit.Title().String(), TargetAmount: habit.Goal().Amount(), Unit: habit.Goal().Unit().String(), Period: habit.Goal().Period().String()}); err != nil {
		return fmt.Errorf("upsert habit: %w", err)
	}
	if err := queries.DeleteNoteTags(ctx, habit.ID().String()); err != nil {
		return err
	}
	for position, tag := range habit.Tags().Strings() {
		if err := queries.EnsureTag(ctx, tag); err != nil {
			return err
		}
		if err := queries.InsertNoteTag(ctx, sqlitedb.InsertNoteTagParams{NoteID: habit.ID().String(), Tag: tag, Position: int64(position)}); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit habit index: %w", err)
	}
	return nil
}

func (index *Index) UpsertPerson(ctx context.Context, person *persons.Person, path string) error {
	if person == nil || person.ID() == uuid.Nil() {
		return fmt.Errorf("initialized person is required")
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("note path is required")
	}
	transaction, err := index.beginWrite(ctx)
	if err != nil {
		return fmt.Errorf("begin person index transaction: %w", err)
	}
	defer transaction.Rollback()
	queries := index.queries.WithTx(transaction.Tx)
	timestamp := func(at time.Time) string { return at.UTC().Format("2006-01-02T15:04:05.000000000Z") }
	if err := queries.UpsertNote(ctx, sqlitedb.UpsertNoteParams{
		ID: person.ID().String(), Type: ports.NoteKindPerson.String(), Path: path,
		CreatedAt: timestamp(person.Metadata().CreatedAt()), UpdatedAt: timestamp(person.Metadata().UpdatedAt()),
	}); err != nil {
		return fmt.Errorf("upsert note: %w", err)
	}
	if err := queries.UpsertPerson(ctx, sqlitedb.UpsertPersonParams{NoteID: person.ID().String(), Name: person.Name().String()}); err != nil {
		return fmt.Errorf("upsert person: %w", err)
	}
	if err := queries.DeletePersonGroups(ctx, person.ID().String()); err != nil {
		return err
	}
	for position, group := range person.Groups().Strings() {
		if err := queries.EnsureGroup(ctx, group); err != nil {
			return err
		}
		if err := queries.InsertPersonGroup(ctx, sqlitedb.InsertPersonGroupParams{PersonID: person.ID().String(), GroupName: group, Position: int64(position)}); err != nil {
			return err
		}
	}
	if err := queries.DeleteNoteTags(ctx, person.ID().String()); err != nil {
		return err
	}
	for position, tag := range person.Tags().Strings() {
		if err := queries.EnsureTag(ctx, tag); err != nil {
			return err
		}
		if err := queries.InsertNoteTag(ctx, sqlitedb.InsertNoteTagParams{NoteID: person.ID().String(), Tag: tag, Position: int64(position)}); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit person index: %w", err)
	}
	return nil
}

func (index *Index) UpsertBookmark(ctx context.Context, entity *bookmarks.Bookmark, path string) error {
	if entity == nil || entity.ID() == uuid.Nil() {
		return fmt.Errorf("initialized entity is required")
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("note path is required")
	}
	transaction, err := index.beginWrite(ctx)
	if err != nil {
		return fmt.Errorf("begin entity index transaction: %w", err)
	}
	defer transaction.Rollback()
	queries := index.queries.WithTx(transaction.Tx)
	timestamp := func(at time.Time) string { return at.UTC().Format("2006-01-02T15:04:05.000000000Z") }
	if err := queries.UpsertNote(ctx, sqlitedb.UpsertNoteParams{
		ID: entity.ID().String(), Type: ports.NoteKindBookmark.String(), Path: path,
		CreatedAt: timestamp(entity.Metadata().CreatedAt()), UpdatedAt: timestamp(entity.Metadata().UpdatedAt()),
	}); err != nil {
		return fmt.Errorf("upsert note: %w", err)
	}
	if err := queries.UpsertBookmark(ctx, sqlitedb.UpsertBookmarkParams{NoteID: entity.ID().String(), Url: entity.URL().String()}); err != nil {
		return fmt.Errorf("upsert bookmark: %w", err)
	}
	if err := queries.DeleteNoteTags(ctx, entity.ID().String()); err != nil {
		return err
	}
	for position, tag := range entity.Tags().Strings() {
		if err := queries.EnsureTag(ctx, tag); err != nil {
			return err
		}
		if err := queries.InsertNoteTag(ctx, sqlitedb.InsertNoteTagParams{NoteID: entity.ID().String(), Tag: tag, Position: int64(position)}); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit entity index: %w", err)
	}
	return nil
}

func (index *Index) UpsertInbox(ctx context.Context, entity *inbox.Inbox, path string) error {
	if entity == nil || entity.ID() == uuid.Nil() {
		return fmt.Errorf("initialized entity is required")
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("note path is required")
	}
	transaction, err := index.beginWrite(ctx)
	if err != nil {
		return fmt.Errorf("begin entity index transaction: %w", err)
	}
	defer transaction.Rollback()
	queries := index.queries.WithTx(transaction.Tx)
	timestamp := func(at time.Time) string { return at.UTC().Format("2006-01-02T15:04:05.000000000Z") }
	if err := queries.UpsertNote(ctx, sqlitedb.UpsertNoteParams{
		ID: entity.ID().String(), Type: ports.NoteKindInbox.String(), Path: path,
		CreatedAt: timestamp(entity.Metadata().CreatedAt()), UpdatedAt: timestamp(entity.Metadata().UpdatedAt()),
	}); err != nil {
		return fmt.Errorf("upsert note: %w", err)
	}
	if err := queries.UpsertInbox(ctx, sqlitedb.UpsertInboxParams{NoteID: entity.ID().String(), Content: entity.Content()}); err != nil {
		return fmt.Errorf("upsert inbox: %w", err)
	}
	if err := queries.DeleteNoteTags(ctx, entity.ID().String()); err != nil {
		return err
	}
	for position, tag := range entity.Tags().Strings() {
		if err := queries.EnsureTag(ctx, tag); err != nil {
			return err
		}
		if err := queries.InsertNoteTag(ctx, sqlitedb.InsertNoteTagParams{NoteID: entity.ID().String(), Tag: tag, Position: int64(position)}); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit entity index: %w", err)
	}
	return nil
}

func (index *Index) Close() error { return index.db.Close() }

func (index *Index) UpsertPage(ctx context.Context, entity *pages.Page, path string) error {
	if entity == nil || entity.ID() == uuid.Nil() {
		return fmt.Errorf("initialized entity is required")
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("note path is required")
	}
	transaction, err := index.beginWrite(ctx)
	if err != nil {
		return fmt.Errorf("begin entity index transaction: %w", err)
	}
	defer transaction.Rollback()
	queries := index.queries.WithTx(transaction.Tx)
	timestamp := func(at time.Time) string { return at.UTC().Format("2006-01-02T15:04:05.000000000Z") }
	if err := queries.UpsertNote(ctx, sqlitedb.UpsertNoteParams{
		ID: entity.ID().String(), Type: ports.NoteKindPage.String(), Path: path,
		CreatedAt: timestamp(entity.Metadata().CreatedAt()), UpdatedAt: timestamp(entity.Metadata().UpdatedAt()),
	}); err != nil {
		return fmt.Errorf("upsert note: %w", err)
	}
	if err := queries.UpsertPage(ctx, sqlitedb.UpsertPageParams{NoteID: entity.ID().String(), Title: entity.Title().String(), Area: entity.Area().String()}); err != nil {
		return fmt.Errorf("upsert page: %w", err)
	}
	if err := queries.DeleteNoteTags(ctx, entity.ID().String()); err != nil {
		return err
	}
	for position, tag := range entity.Tags().Strings() {
		if err := queries.EnsureTag(ctx, tag); err != nil {
			return err
		}
		if err := queries.InsertNoteTag(ctx, sqlitedb.InsertNoteTagParams{NoteID: entity.ID().String(), Tag: tag, Position: int64(position)}); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit entity index: %w", err)
	}
	return nil
}

func (index *Index) UpsertJournal(ctx context.Context, entity *journals.Journal, path string) error {
	if entity == nil || entity.ID() == uuid.Nil() {
		return fmt.Errorf("initialized entity is required")
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("note path is required")
	}
	transaction, err := index.beginWrite(ctx)
	if err != nil {
		return fmt.Errorf("begin entity index transaction: %w", err)
	}
	defer transaction.Rollback()
	queries := index.queries.WithTx(transaction.Tx)
	timestamp := func(at time.Time) string { return at.UTC().Format("2006-01-02T15:04:05.000000000Z") }
	if err := queries.UpsertNote(ctx, sqlitedb.UpsertNoteParams{
		ID: entity.ID().String(), Type: ports.NoteKindJournal.String(), Path: path,
		CreatedAt: timestamp(entity.Metadata().CreatedAt()), UpdatedAt: timestamp(entity.Metadata().UpdatedAt()),
	}); err != nil {
		return fmt.Errorf("upsert note: %w", err)
	}
	if err := queries.UpsertJournal(ctx, sqlitedb.UpsertJournalParams{NoteID: entity.ID().String(), Date: entity.Date().String()}); err != nil {
		var sqliteError *sqlite.Error
		if errors.As(err, &sqliteError) && sqliteError.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			return fmt.Errorf("%w: %s", ports.ErrJournalAlreadyExists, entity.Date())
		}
		return fmt.Errorf("upsert journal: %w", err)
	}
	if err := queries.DeleteNoteTags(ctx, entity.ID().String()); err != nil {
		return err
	}
	for position, tag := range entity.Tags().Strings() {
		if err := queries.EnsureTag(ctx, tag); err != nil {
			return err
		}
		if err := queries.InsertNoteTag(ctx, sqlitedb.InsertNoteTagParams{NoteID: entity.ID().String(), Tag: tag, Position: int64(position)}); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit entity index: %w", err)
	}
	return nil
}

func (index *Index) JournalExists(ctx context.Context, date calendar.Date) (bool, error) {
	if date.IsZero() {
		return false, fmt.Errorf("journal date is required")
	}
	return index.queries.JournalExists(ctx, date.String())
}
