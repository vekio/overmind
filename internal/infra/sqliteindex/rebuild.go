package sqliteindex

import (
	"context"
	"fmt"
	"time"

	"github.com/vekio/overmind/internal/infra/sqliteindex/sqlitedb"
	"github.com/vekio/overmind/internal/ports"
)

// ReplaceAll atomically replaces every indexed note with records.
func (store *Store) ReplaceAll(ctx context.Context, records []ports.IndexRecord) error {
	transaction, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin index rebuild: %w", err)
	}
	defer transaction.Rollback()

	if _, err := transaction.ExecContext(ctx, "DELETE FROM notes"); err != nil {
		return fmt.Errorf("clear index: %w", err)
	}
	queries := store.queries.WithTx(transaction)
	for _, record := range records {
		if err := insertRecord(ctx, queries, record); err != nil {
			return fmt.Errorf("index note %s: %w", record.ID, err)
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit index rebuild: %w", err)
	}
	return nil
}

func insertRecord(ctx context.Context, queries *sqlitedb.Queries, record ports.IndexRecord) error {
	id := record.ID.String()
	if err := queries.UpsertNote(ctx, sqlitedb.UpsertNoteParams{
		ID:        id,
		Kind:      record.Kind.String(),
		CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: record.UpdatedAt.UTC().Format(time.RFC3339),
	}); err != nil {
		return err
	}
	for _, attribute := range record.Attributes {
		if err := queries.InsertNoteAttribute(ctx, sqlitedb.InsertNoteAttributeParams{
			NoteID: id,
			Name:   attribute.Name,
			Value:  attribute.Value,
		}); err != nil {
			return err
		}
	}
	for position, tag := range record.Tags {
		if err := queries.InsertNoteTag(ctx, sqlitedb.InsertNoteTagParams{
			NoteID:   id,
			Tag:      tag,
			Position: int64(position),
		}); err != nil {
			return err
		}
	}
	return nil
}
