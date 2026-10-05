package bootstrap

import (
	"context"
	"io"
	"path/filepath"

	"github.com/vekio/overmind/internal/app"
	appconfig "github.com/vekio/overmind/internal/config"
	"github.com/vekio/overmind/internal/infra/codecs"
	"github.com/vekio/overmind/internal/infra/idgenerator"
	"github.com/vekio/overmind/internal/infra/index"
	"github.com/vekio/overmind/internal/infra/notestore"
	"github.com/vekio/overmind/internal/infra/repositories"
)

// buildApplication wires all use cases to a shared vault store and SQLite index.
// The returned resources belong to the runtime and must be closed after use.
func buildApplication(ctx context.Context, settings appconfig.Settings) (*app.Application, []io.Closer, error) {
	noteIndex, err := index.New(ctx, filepath.Join(settings.VaultPath, "index.db"))
	if err != nil {
		return nil, nil, err
	}
	notes := notestore.New(filepath.Join(settings.VaultPath, "notes"))
	application := app.New(app.Dependencies{
		IDGenerator:  idgenerator.New(),
		Index:        noteIndex,
		RawNoteCodec: codecs.RawNoteCodec{},
		NoteStore:    notes,
		Habits:       repositories.NewHabitRepository(notes, noteIndex, codecs.HabitCodec{}),
		Persons:      repositories.NewPersonRepository(notes, noteIndex, codecs.PersonCodec{}),
		Bookmarks:    repositories.NewBookmarkRepository(notes, noteIndex, codecs.BookmarkCodec{}),
		Inbox:        repositories.NewInboxRepository(notes, noteIndex, codecs.InboxCodec{}),
		Pages:        repositories.NewPageRepository(notes, noteIndex, codecs.PageCodec{}),
		Journals:     repositories.NewJournalRepository(notes, noteIndex, codecs.JournalCodec{}),
	})
	return application, []io.Closer{noteIndex}, nil
}
