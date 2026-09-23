// Package sqliteindex indexes note metadata in SQLite.
package sqliteindex

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/infra/sqliteindex/sqlitedb"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

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

	transaction, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin note index transaction: %w", err)
	}
	defer transaction.Rollback()

	queries := store.queries.WithTx(transaction)
	metadata := note.Metadata()
	if err := queries.UpsertNote(ctx, sqlitedb.UpsertNoteParams{
		ID:        metadata.ID().String(),
		Kind:      metadata.Kind().String(),
		CreatedAt: metadata.CreatedAt().UTC().Format(time.RFC3339),
		UpdatedAt: metadata.UpdatedAt().UTC().Format(time.RFC3339),
	}); err != nil {
		return fmt.Errorf("upsert indexed note: %w", err)
	}
	if err := queries.DeleteNoteAttributes(ctx, metadata.ID().String()); err != nil {
		return fmt.Errorf("replace indexed note attributes: %w", err)
	}
	for _, attribute := range attributes {
		if err := queries.InsertNoteAttribute(ctx, sqlitedb.InsertNoteAttributeParams{
			NoteID: metadata.ID().String(),
			Name:   attribute.name,
			Value:  attribute.value,
		}); err != nil {
			return fmt.Errorf("insert indexed note attribute %q: %w", attribute.name, err)
		}
	}
	if err := queries.DeleteNoteTags(ctx, metadata.ID().String()); err != nil {
		return fmt.Errorf("replace indexed note tags: %w", err)
	}
	for position, tag := range metadata.Tags().Strings() {
		if err := queries.InsertNoteTag(ctx, sqlitedb.InsertNoteTagParams{
			NoteID:   metadata.ID().String(),
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

type noteAttribute struct {
	name  string
	value string
}

func attributesFromNote(note domain.Note) ([]noteAttribute, error) {
	switch note := note.(type) {
	case domain.Page:
		attributes := []noteAttribute{{name: "title", value: note.Title().String()}}
		if !note.Area().IsZero() {
			attributes = append(attributes, noteAttribute{name: "area", value: note.Area().String()})
		}
		return attributes, nil
	case domain.Journal:
		return []noteAttribute{{name: "date", value: note.Date().String()}}, nil
	case domain.Inbox:
		return nil, nil
	case domain.Bookmark:
		return []noteAttribute{{name: "url", value: note.URL().String()}}, nil
	default:
		return nil, fmt.Errorf("index unsupported note type %T", note)
	}
}

// Close closes the SQLite database.
func (store *Store) Close() error {
	return store.db.Close()
}
