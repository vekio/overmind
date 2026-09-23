// Package app implements Overmind use cases.
package app

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/internal/domain"
	"uuid"
)

// NoteRenderer renders a note as a document.
type NoteRenderer interface {
	Render(context.Context, domain.Note) ([]byte, error)
}

// NoteWriter persists a rendered note.
type NoteWriter interface {
	Write(context.Context, uuid.UUID, []byte) (string, error)
}

// NoteIndex indexes note metadata.
type NoteIndex interface {
	Upsert(context.Context, domain.Note) error
}

// App provides the application use cases.
type App struct {
	renderer NoteRenderer
	writer   NoteWriter
	index    NoteIndex
}

// New creates the application use cases.
func New(renderer NoteRenderer, writer NoteWriter, index NoteIndex) *App {
	return &App{
		renderer: renderer,
		writer:   writer,
		index:    index,
	}
}

func (app *App) save(ctx context.Context, note domain.Note) (string, error) {
	content, err := app.renderer.Render(ctx, note)
	if err != nil {
		return "", err
	}

	path, err := app.writer.Write(ctx, note.Metadata().ID(), content)
	if err != nil {
		return "", err
	}
	if err := app.index.Upsert(ctx, note); err != nil {
		return "", fmt.Errorf("index note: %w", err)
	}

	return path, nil
}
