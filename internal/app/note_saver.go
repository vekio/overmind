package app

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

type noteSaver struct {
	renderer ports.NoteRenderer
	writer   ports.NoteWriter
	index    ports.NoteIndex
}

func newNoteSaver(renderer ports.NoteRenderer, writer ports.NoteWriter, index ports.NoteIndex) *noteSaver {
	return &noteSaver{renderer: renderer, writer: writer, index: index}
}

func (saver *noteSaver) save(ctx context.Context, note domain.Note) (string, error) {
	content, err := saver.renderer.Render(ctx, note)
	if err != nil {
		return "", err
	}

	path, err := saver.writer.Write(ctx, note.Metadata().ID(), content)
	if err != nil {
		return "", err
	}
	if err := saver.index.Upsert(ctx, note); err != nil {
		return "", fmt.Errorf("index note: %w", err)
	}

	return path, nil
}
