package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
	"git.casta.me/alberto/overmind/pkg/asciidoc/edit"
)

// ErrNoteChanged reports that the stored document differs from the opened version.
var ErrNoteChanged = errors.New("note changed since it was opened")

// UpdateNoteCommand contains the new document and the source originally opened.
type UpdateNoteCommand struct {
	ID       uuid.UUID
	Original []byte
	Source   []byte
}

// UpdateNoteResult returns the exact persisted source, including its new timestamp.
type UpdateNoteResult struct {
	Path   string
	Source []byte
}

// UpdateNoteHandler validates and saves a complete edited AsciiDoc document.
type UpdateNoteHandler struct {
	reader ports.NoteReader
	writer ports.NoteWriter
	parser ports.NoteParser
	index  ports.NoteIndex
}

func newUpdateNoteHandler(reader ports.NoteReader, writer ports.NoteWriter, parser ports.NoteParser, index ports.NoteIndex) *UpdateNoteHandler {
	return &UpdateNoteHandler{reader: reader, writer: writer, parser: parser, index: index}
}

func (handler *UpdateNoteHandler) Handle(ctx context.Context, command UpdateNoteCommand) (UpdateNoteResult, error) {
	if command.ID == uuid.Nil() {
		return UpdateNoteResult{}, fmt.Errorf("note ID is required")
	}
	current, err := handler.reader.Read(ctx, command.ID)
	if err != nil {
		return UpdateNoteResult{}, fmt.Errorf("read note before update: %w", err)
	}
	if !bytes.Equal(current, command.Original) {
		return UpdateNoteResult{}, ErrNoteChanged
	}
	previous, err := handler.parser.Parse(current)
	if err != nil {
		return UpdateNoteResult{}, fmt.Errorf("parse current note: %w", err)
	}
	if previous.ID != command.ID {
		return UpdateNoteResult{}, fmt.Errorf("stored note ID does not match %s", command.ID)
	}

	editor, err := edit.New(command.Source)
	if err != nil {
		return UpdateNoteResult{}, fmt.Errorf("parse edited note: %w", err)
	}
	updatedAt := time.Now().UTC().Truncate(time.Second)
	if !updatedAt.After(previous.UpdatedAt) {
		updatedAt = previous.UpdatedAt.Add(time.Second)
	}
	if err := editor.SetHeaderAttribute("overmind-updated-at", updatedAt.Format(time.RFC3339)); err != nil {
		return UpdateNoteResult{}, fmt.Errorf("update note timestamp: %w", err)
	}
	source := editor.Bytes()
	record, err := handler.parser.Parse(source)
	if err != nil {
		return UpdateNoteResult{}, fmt.Errorf("parse edited note metadata: %w", err)
	}
	if record.ID != command.ID || record.Kind != previous.Kind || !record.CreatedAt.Equal(previous.CreatedAt) {
		return UpdateNoteResult{}, fmt.Errorf("edited note changes its identity or creation time")
	}
	if record.Kind == domain.NoteKindJournal && attribute(record.Attributes, "date") != attribute(previous.Attributes, "date") {
		return UpdateNoteResult{}, fmt.Errorf("edited journal changes its date")
	}
	path, err := handler.writer.Write(ctx, command.ID, source)
	if err != nil {
		return UpdateNoteResult{}, fmt.Errorf("write edited note: %w", err)
	}
	if err := handler.index.UpsertRecord(ctx, record); err != nil {
		return UpdateNoteResult{}, fmt.Errorf("index edited note: %w", err)
	}
	return UpdateNoteResult{Path: path, Source: source}, nil
}

func attribute(attributes []ports.IndexAttribute, name string) string {
	for _, item := range attributes {
		if item.Name == name {
			return item.Value
		}
	}
	return ""
}
