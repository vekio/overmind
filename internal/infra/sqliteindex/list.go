package sqliteindex

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain"
	"github.com/vekio/overmind/internal/infra/sqliteindex/sqlitedb"
	"github.com/vekio/overmind/internal/ports"
)

var _ ports.NoteLister = (*Store)(nil)

// List returns indexed notes with attributes and tags, newest update first.
func (store *Store) List(ctx context.Context) ([]ports.IndexRecord, error) {
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin note list: %w", err)
	}
	defer tx.Rollback()
	queries := store.queries.WithTx(tx)

	notes, err := queries.ListNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("read indexed notes: %w", err)
	}
	attributes, err := queries.ListNoteAttributes(ctx)
	if err != nil {
		return nil, fmt.Errorf("read indexed note attributes: %w", err)
	}
	tags, err := queries.ListNoteTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("read indexed note tags: %w", err)
	}

	records := make([]ports.IndexRecord, 0, len(notes))
	positions := make(map[string]int, len(notes))
	for _, note := range notes {
		record, err := listedRecord(note)
		if err != nil {
			return nil, err
		}
		positions[note.ID] = len(records)
		records = append(records, record)
	}
	for _, attribute := range attributes {
		position, ok := positions[attribute.NoteID]
		if !ok {
			return nil, fmt.Errorf("orphan indexed attribute for note %q", attribute.NoteID)
		}
		records[position].Attributes = append(records[position].Attributes, ports.IndexAttribute{Name: attribute.Name, Value: attribute.Value})
	}
	for _, tag := range tags {
		position, ok := positions[tag.NoteID]
		if !ok {
			return nil, fmt.Errorf("orphan indexed tag for note %q", tag.NoteID)
		}
		records[position].Tags = append(records[position].Tags, tag.Tag)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("finish note list: %w", err)
	}
	return records, nil
}

func listedRecord(note sqlitedb.Note) (ports.IndexRecord, error) {
	id, err := uuid.Parse(note.ID)
	if err != nil {
		return ports.IndexRecord{}, fmt.Errorf("invalid indexed note ID %q: %w", note.ID, err)
	}
	kind, err := domain.ParseNoteKind(note.Kind)
	if err != nil {
		return ports.IndexRecord{}, fmt.Errorf("invalid indexed note %s: %w", id, err)
	}
	createdAt, err := time.Parse(time.RFC3339, note.CreatedAt)
	if err != nil {
		return ports.IndexRecord{}, fmt.Errorf("invalid creation time for indexed note %s: %w", id, err)
	}
	updatedAt, err := time.Parse(time.RFC3339, note.UpdatedAt)
	if err != nil {
		return ports.IndexRecord{}, fmt.Errorf("invalid update time for indexed note %s: %w", id, err)
	}
	return ports.IndexRecord{ID: id, Kind: kind, CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}
