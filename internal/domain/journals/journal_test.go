package journals_test

import (
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/domain/shared"
)

func TestJournalEditsKeepDateAndRejectBackwardTimes(t *testing.T) {
	at := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	metadata, err := shared.NewEntityMetadata(at, at)
	if err != nil {
		t.Fatal(err)
	}
	date, err := calendar.NewDate("2026-10-05")
	if err != nil {
		t.Fatal(err)
	}
	tag, err := shared.NewTag("daily")
	if err != nil {
		t.Fatal(err)
	}
	tags, err := shared.NewTags(tag)
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	entity, err := journals.NewJournal(id, date, "Original", tags, metadata)
	if err != nil {
		t.Fatal(err)
	}
	later := at.Add(time.Hour)
	body := "\n== Reflection\nText with trailing spaces  \n"
	if err := entity.Rewrite(body, later); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []func() error{
		func() error { return entity.Rewrite("Rejected", at) },
		func() error { return entity.ReplaceTags(shared.Tags{}, at) },
	} {
		if err := mutation(); !errors.Is(err, shared.ErrInvalidEntityMetadata) {
			t.Fatalf("backward mutation = %v", err)
		}
		if entity.Content() != body || !entity.Tags().Contains(tag) || !entity.Metadata().UpdatedAt().Equal(later) {
			t.Fatal("failed mutation changed journal content, tags or metadata")
		}
	}
	if err := entity.Rewrite("", later); err != nil {
		t.Fatal(err)
	}
	if err := entity.ReplaceTags(shared.Tags{}, later); err != nil {
		t.Fatal(err)
	}
	if entity.ID() != id || !entity.Date().Equal(date) || !entity.Metadata().CreatedAt().Equal(at) || entity.Content() != "" || !entity.Tags().IsEmpty() {
		t.Fatal("edits changed fixed journal fields or failed to clear optional fields")
	}
}

func TestJournalRequiresInitializedDateIdentityAndMetadata(t *testing.T) {
	at := time.Now()
	metadata, err := shared.NewEntityMetadata(at, at)
	if err != nil {
		t.Fatal(err)
	}
	date, err := calendar.NewDate("2026-10-05")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		id       uuid.UUID
		date     calendar.Date
		metadata shared.EntityMetadata
	}{
		{uuid.Nil(), date, metadata},
		{uuid.New(), calendar.Date{}, metadata},
		{uuid.New(), date, shared.EntityMetadata{}},
	} {
		if _, err := journals.NewJournal(tc.id, tc.date, "", shared.Tags{}, tc.metadata); !errors.Is(err, journals.ErrInvalidJournal) {
			t.Fatalf("invalid journal = %v", err)
		}
	}
	var uninitialized *journals.Journal
	if err := uninitialized.ReplaceTags(shared.Tags{}, at); !errors.Is(err, journals.ErrInvalidJournal) {
		t.Fatalf("nil receiver mutation = %v", err)
	}
}
