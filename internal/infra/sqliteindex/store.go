// Package sqliteindex indexes note metadata in SQLite.
package sqliteindex

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"uuid"

	"github.com/pressly/goose/v3"
	"github.com/vekio/overmind/internal/domain"
	"github.com/vekio/overmind/internal/infra/sqliteindex/sqlitedb"
	"github.com/vekio/overmind/internal/ports"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

var _ ports.NoteIndex = (*Store)(nil)

// Store is a SQLite note index.
type Store struct {
	db      *sql.DB
	queries *sqlitedb.Queries
}

// New opens an index and applies its migrations.
func New(ctx context.Context, path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("SQLite index path is required")
	}
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("create SQLite index directory: %w", err)
		}
	}

	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open SQLite index: %w", err)
	}
	database.SetMaxOpenConns(1)
	if err := configure(ctx, database); err != nil {
		_ = database.Close()
		return nil, err
	}
	if err := migrate(ctx, database); err != nil {
		_ = database.Close()
		return nil, err
	}

	return &Store{
		db:      database,
		queries: sqlitedb.New(database),
	}, nil
}

func configure(ctx context.Context, database *sql.DB) error {
	for _, statement := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA journal_mode = WAL",
	} {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("configure SQLite index: %w", err)
		}
	}

	return nil
}

func migrate(ctx context.Context, database *sql.DB) error {
	migrations, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("open SQLite index migrations: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, database, migrations)
	if err != nil {
		return fmt.Errorf("create SQLite migration provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("migrate SQLite index: %w", err)
	}

	return nil
}

// Upsert adds or replaces a note and its tags atomically.
func (store *Store) Upsert(ctx context.Context, note domain.Note) error {
	attributes, err := attributesFromNote(note)
	if err != nil {
		return err
	}
	metadata := note.Metadata()
	return store.UpsertRecord(ctx, ports.IndexRecord{
		ID: metadata.ID(), Kind: metadata.Kind(),
		CreatedAt: metadata.CreatedAt(), UpdatedAt: metadata.UpdatedAt(),
		Attributes: attributes, Tags: metadata.Tags().Strings(),
	})
}

// UpsertRecord adds or replaces indexed metadata for an edited source document.
func (store *Store) UpsertRecord(ctx context.Context, record ports.IndexRecord) error {
	transaction, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin note index transaction: %w", err)
	}
	defer transaction.Rollback()

	queries := store.queries.WithTx(transaction)
	if err := queries.UpsertNote(ctx, sqlitedb.UpsertNoteParams{
		ID:        record.ID.String(),
		Kind:      record.Kind.String(),
		CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: record.UpdatedAt.UTC().Format(time.RFC3339),
	}); err != nil {
		return fmt.Errorf("upsert indexed note: %w", err)
	}
	if err := queries.DeleteNoteAttributes(ctx, record.ID.String()); err != nil {
		return fmt.Errorf("replace indexed note attributes: %w", err)
	}
	for _, attribute := range record.Attributes {
		if err := queries.InsertNoteAttribute(ctx, sqlitedb.InsertNoteAttributeParams{
			NoteID: record.ID.String(),
			Name:   attribute.Name,
			Value:  attribute.Value,
		}); err != nil {
			return fmt.Errorf("insert indexed note attribute %q: %w", attribute.Name, err)
		}
	}
	if err := queries.DeleteNoteTags(ctx, record.ID.String()); err != nil {
		return fmt.Errorf("replace indexed note tags: %w", err)
	}
	for position, tag := range record.Tags {
		if err := queries.InsertNoteTag(ctx, sqlitedb.InsertNoteTagParams{
			NoteID:   record.ID.String(),
			Tag:      tag,
			Position: int64(position),
		}); err != nil {
			return fmt.Errorf("insert indexed note tag %q: %w", tag, err)
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit note index transaction: %w", err)
	}

	return nil
}

// Delete removes a note and its attributes and tags from the index.
func (store *Store) Delete(ctx context.Context, id uuid.UUID) error {
	if err := store.queries.DeleteNote(ctx, id.String()); err != nil {
		return fmt.Errorf("delete indexed note %s: %w", id, err)
	}
	return nil
}

// JournalExists reports whether a journal is indexed for date.
func (store *Store) JournalExists(ctx context.Context, date domain.Date) (bool, error) {
	exists, err := store.queries.JournalExists(ctx, date.String())
	if err != nil {
		return false, fmt.Errorf("check indexed journal date: %w", err)
	}

	return exists, nil
}

// JournalID finds the indexed journal for a date.
func (store *Store) JournalID(ctx context.Context, date domain.Date) (uuid.UUID, bool, error) {
	value, err := store.queries.JournalID(ctx, date.String())
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil(), false, nil
	}
	if err != nil {
		return uuid.Nil(), false, fmt.Errorf("find indexed journal: %w", err)
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil(), false, fmt.Errorf("invalid indexed journal ID %q: %w", value, err)
	}
	return id, true, nil
}

func attributesFromNote(note domain.Note) ([]ports.IndexAttribute, error) {
	switch note := note.(type) {
	case domain.Person:
		attributes := []ports.IndexAttribute{{Name: "name", Value: note.Name().String()}}
		if !note.Groups().IsEmpty() {
			attributes = append(attributes, ports.IndexAttribute{Name: "groups", Value: strings.Join(note.Groups().Strings(), ", ")})
		}
		return attributes, nil
	case domain.Page:
		attributes := []ports.IndexAttribute{{Name: "title", Value: note.Title().String()}}
		if !note.Area().IsZero() {
			attributes = append(attributes, ports.IndexAttribute{Name: "area", Value: note.Area().String()})
		}
		return attributes, nil
	case domain.Journal:
		return []ports.IndexAttribute{{Name: "date", Value: note.Date().String()}}, nil
	case domain.Inbox:
		return nil, nil
	case domain.Bookmark:
		return []ports.IndexAttribute{{Name: "url", Value: note.URL().String()}}, nil
	default:
		return nil, fmt.Errorf("index unsupported note type %T", note)
	}
}

// Close closes the SQLite database.
func (store *Store) Close() error {
	return store.db.Close()
}
