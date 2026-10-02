package app

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/ports"
)

// RebuildIndexResult reports how many notes were indexed.
type RebuildIndexResult struct {
	Count int
}

// RebuildIndexHandler reconstructs the index from note documents.
type RebuildIndexHandler struct {
	walker ports.NoteWalker
	parser ports.NoteParser
	index  ports.NoteIndex
}

func newRebuildIndexHandler(walker ports.NoteWalker, parser ports.NoteParser, index ports.NoteIndex) *RebuildIndexHandler {
	return &RebuildIndexHandler{walker: walker, parser: parser, index: index}
}

// Handle parses every note before atomically replacing the index.
func (handler *RebuildIndexHandler) Handle(ctx context.Context) (RebuildIndexResult, error) {
	var records []ports.IndexRecord
	seen := make(map[uuid.UUID]string)
	err := handler.walker.Walk(ctx, func(path string, content []byte) error {
		record, err := handler.parser.Parse(content)
		if err != nil {
			return fmt.Errorf("parse note %q: %w", path, err)
		}
		if previous, exists := seen[record.ID]; exists {
			return fmt.Errorf("duplicate note ID %s in %q and %q", record.ID, previous, path)
		}
		seen[record.ID] = path
		records = append(records, record)
		return nil
	})
	if err != nil {
		return RebuildIndexResult{}, err
	}
	if err := handler.index.ReplaceAll(ctx, records); err != nil {
		return RebuildIndexResult{}, fmt.Errorf("rebuild index: %w", err)
	}
	return RebuildIndexResult{Count: len(records)}, nil
}
