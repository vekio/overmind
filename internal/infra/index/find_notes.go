package index

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/infra/index/sqlitedb"
	"github.com/vekio/overmind/internal/ports"
)

var _ ports.NoteFinder = (*Index)(nil)

func (index *Index) FindNotes(ctx context.Context, filter ports.NoteFilter) ([]ports.NoteSummary, error) {
	rows, err := index.queries.FindNotes(ctx, sqlitedb.FindNotesParams{
		NoteType: filter.Type, Tag: filter.Tag, PageLimit: int64(filter.Limit), PageOffset: int64(filter.Offset),
	})
	if err != nil {
		return nil, err
	}
	notes := make([]ports.NoteSummary, 0, len(rows))
	for _, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("parse indexed ID: %w", err)
		}
		at, err := time.Parse(time.RFC3339Nano, row.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse indexed timestamp: %w", err)
		}
		var tags []string
		if err := json.Unmarshal([]byte(row.Tags), &tags); err != nil {
			return nil, fmt.Errorf("parse indexed tags: %w", err)
		}
		// Display labels are single-line; the document and stored content stay intact.
		label := strings.Join(strings.Fields(row.Label), " ")
		notes = append(notes, ports.NoteSummary{ID: id, Type: row.Type, Label: label, Tags: tags, UpdatedAt: at})
	}
	return notes, nil
}
