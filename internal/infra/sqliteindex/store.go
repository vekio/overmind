// Package sqliteindex implements the document read model with SQLite.
package sqliteindex

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/infra/sqliteindex/sqlitedb"
	"git.casta.me/alberto/overmind/internal/ports"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

var _ ports.DocumentIndex = (*Store)(nil)

// Store is a SQLite-backed document index.
type Store struct {
	db      *sql.DB
	queries *sqlitedb.Queries
}

// New opens the index and applies all embedded migrations.
func New(ctx context.Context, databasePath string) (*Store, error) {
	if databasePath == "" {
		return nil, fmt.Errorf("SQLite index path is required")
	}
	if databasePath != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
			return nil, fmt.Errorf("create SQLite index directory: %w", err)
		}
	}

	database, err := sql.Open("sqlite", databasePath)
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

	return &Store{db: database, queries: sqlitedb.New(database)}, nil
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

// Close closes the underlying database.
func (store *Store) Close() error {
	return store.db.Close()
}

// GetByID returns a document and all its Overmind attributes.
func (store *Store) GetByID(ctx context.Context, id domain.DocumentID) (ports.IndexedDocument, error) {
	document, err := store.queries.GetDocumentByID(ctx, id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return ports.IndexedDocument{}, ports.ErrIndexedDocumentNotFound
	}
	if err != nil {
		return ports.IndexedDocument{}, fmt.Errorf("get indexed document: %w", err)
	}
	kind, err := domain.NewDocumentKind(document.Kind)
	if err != nil {
		return ports.IndexedDocument{}, fmt.Errorf("decode indexed document kind: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339, document.CreatedAt)
	if err != nil {
		return ports.IndexedDocument{}, fmt.Errorf("decode indexed document creation time: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, document.UpdatedAt)
	if err != nil {
		return ports.IndexedDocument{}, fmt.Errorf("decode indexed document update time: %w", err)
	}
	attributes, err := store.queries.ListDocumentAttributes(ctx, id.String())
	if err != nil {
		return ports.IndexedDocument{}, fmt.Errorf("list indexed document attributes: %w", err)
	}
	tags, err := store.queries.ListDocumentTags(ctx, id.String())
	if err != nil {
		return ports.IndexedDocument{}, fmt.Errorf("list indexed document tags: %w", err)
	}

	result := ports.IndexedDocument{
		ID:         id,
		Kind:       kind,
		Title:      document.Title,
		Area:       document.Area,
		Tags:       tags,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Attributes: make(map[string]string, len(attributes)),
	}
	for _, attribute := range attributes {
		result.Attributes[attribute.Name] = attribute.Value
	}
	return result, nil
}

// List returns lightweight document summaries matching every filter.
func (store *Store) List(
	ctx context.Context,
	filter ports.ListIndexedDocumentsFilter,
) ([]ports.IndexedDocumentSummary, error) {
	tagsJSON, err := json.Marshal(filter.Tags.Strings())
	if err != nil {
		return nil, fmt.Errorf("encode indexed document tag filter: %w", err)
	}
	indexed, err := store.queries.ListDocuments(ctx, sqlitedb.ListDocumentsParams{
		Kind:     filter.Kind.String(),
		Title:    filter.Title,
		Area:     filter.Area.String(),
		TagsJson: string(tagsJSON),
	})
	if err != nil {
		return nil, fmt.Errorf("list indexed documents: %w", err)
	}

	documents := make([]ports.IndexedDocumentSummary, len(indexed))
	for index, document := range indexed {
		id, err := domain.NewDocumentID(document.ID)
		if err != nil {
			return nil, fmt.Errorf("decode indexed document id: %w", err)
		}
		kind, err := domain.NewDocumentKind(document.Kind)
		if err != nil {
			return nil, fmt.Errorf("decode indexed document kind: %w", err)
		}
		tags, err := store.queries.ListDocumentTags(ctx, document.ID)
		if err != nil {
			return nil, fmt.Errorf("list indexed document tags: %w", err)
		}
		documents[index] = ports.IndexedDocumentSummary{
			ID: id, Kind: kind, Title: document.Title, Area: document.Area, Tags: tags,
		}
	}
	return documents, nil
}

// Upsert replaces one indexed document and its attributes atomically.
func (store *Store) Upsert(ctx context.Context, document ports.IndexedDocument) error {
	transaction, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin index transaction: %w", err)
	}
	defer transaction.Rollback()

	if err := upsert(ctx, store.queries.WithTx(transaction), document); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit index transaction: %w", err)
	}
	return nil
}

// ReplaceAll atomically replaces the complete document read model.
func (store *Store) ReplaceAll(ctx context.Context, documents []ports.IndexedDocument) error {
	transaction, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin index rebuild: %w", err)
	}
	defer transaction.Rollback()

	queries := store.queries.WithTx(transaction)
	if err := queries.DeleteAllDocuments(ctx); err != nil {
		return fmt.Errorf("clear document index: %w", err)
	}
	for _, document := range documents {
		if err := upsert(ctx, queries, document); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit index rebuild: %w", err)
	}
	return nil
}

func upsert(ctx context.Context, queries *sqlitedb.Queries, document ports.IndexedDocument) error {
	if err := queries.UpsertDocument(ctx, sqlitedb.UpsertDocumentParams{
		ID:        document.ID.String(),
		Kind:      document.Kind.String(),
		Title:     document.Title,
		Area:      document.Area,
		CreatedAt: document.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: document.UpdatedAt.UTC().Format(time.RFC3339),
	}); err != nil {
		return fmt.Errorf("upsert indexed document: %w", err)
	}
	if err := queries.DeleteDocumentAttributes(ctx, document.ID.String()); err != nil {
		return fmt.Errorf("replace indexed document attributes: %w", err)
	}
	if err := queries.DeleteDocumentTags(ctx, document.ID.String()); err != nil {
		return fmt.Errorf("replace indexed document tags: %w", err)
	}
	for position, tag := range document.Tags {
		if err := queries.InsertDocumentTag(ctx, sqlitedb.InsertDocumentTagParams{
			DocumentID: document.ID.String(),
			Tag:        tag,
			Position:   int64(position),
		}); err != nil {
			return fmt.Errorf("insert indexed document tag %q: %w", tag, err)
		}
	}

	names := make([]string, 0, len(document.Attributes))
	for name := range document.Attributes {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		if err := queries.InsertDocumentAttribute(ctx, sqlitedb.InsertDocumentAttributeParams{
			DocumentID: document.ID.String(),
			Name:       name,
			Value:      document.Attributes[name],
		}); err != nil {
			return fmt.Errorf("insert indexed document attribute %q: %w", name, err)
		}
	}
	return nil
}
