// Package projections maps note documents to their indexed domain projections.
package projections

import (
	"context"
	"fmt"

	"github.com/vekio/overmind/internal/infra/codecs"
	"github.com/vekio/overmind/internal/ports"
)

// NoteProjector maps managed AsciiDoc documents to validated index projections.
type NoteProjector struct{}

var _ ports.NoteProjector = NoteProjector{}

func (NoteProjector) Project(ctx context.Context, document ports.Note, index ports.Index) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	id, kind, err := codecs.Identify(document.Content)
	if err != nil {
		return err
	}
	if id != document.ID {
		return fmt.Errorf("document ID %s does not match filename ID %s", id, document.ID)
	}
	switch kind {
	case ports.NoteKindHabit:
		entity, err := (codecs.HabitCodec{}).Decode(document.Content)
		if err != nil {
			return err
		}
		return index.UpsertHabit(ctx, entity, document.Path)
	case ports.NoteKindPerson:
		entity, err := (codecs.PersonCodec{}).Decode(document.Content)
		if err != nil {
			return err
		}
		return index.UpsertPerson(ctx, entity, document.Path)
	case ports.NoteKindBookmark:
		entity, err := (codecs.BookmarkCodec{}).Decode(document.Content)
		if err != nil {
			return err
		}
		return index.UpsertBookmark(ctx, entity, document.Path)
	case ports.NoteKindInbox:
		entity, err := (codecs.InboxCodec{}).Decode(document.Content)
		if err != nil {
			return err
		}
		return index.UpsertInbox(ctx, entity, document.Path)
	case ports.NoteKindPage:
		entity, err := (codecs.PageCodec{}).Decode(document.Content)
		if err != nil {
			return err
		}
		return index.UpsertPage(ctx, entity, document.Path)
	case ports.NoteKindJournal:
		entity, err := (codecs.JournalCodec{}).Decode(document.Content)
		if err != nil {
			return err
		}
		return index.UpsertJournal(ctx, entity, document.Path)
	default:
		return fmt.Errorf("unsupported note type %q", kind)
	}
}
