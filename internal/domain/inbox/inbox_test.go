package inbox_test

import (
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/domain/shared"
)

func TestInboxEditsPreserveIdentityAndRejectBackwardTimes(t *testing.T) {
	at := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	metadata, err := shared.NewEntityMetadata(at, at)
	if err != nil {
		t.Fatal(err)
	}
	tag, err := shared.NewTag("pending")
	if err != nil {
		t.Fatal(err)
	}
	tags, err := shared.NewTags(tag)
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	entity, err := inbox.NewInbox(id, "Original", tags, metadata)
	if err != nil {
		t.Fatal(err)
	}
	later := at.Add(time.Hour)
	body := "\n== Content\r\nText with trailing spaces  \r\n"
	if err := entity.Rewrite(body, later); err != nil {
		t.Fatal(err)
	}
	if entity.Content() != body || entity.ID() != id || !entity.Metadata().CreatedAt().Equal(at) || !entity.Metadata().UpdatedAt().Equal(later) {
		t.Fatal("rewrite changed body bytes, identity or lifecycle metadata")
	}
	for _, mutation := range []func() error{
		func() error { return entity.Rewrite("Rejected", at) },
		func() error { return entity.ReplaceTags(shared.Tags{}, at) },
	} {
		if err := mutation(); !errors.Is(err, shared.ErrInvalidEntityMetadata) {
			t.Fatalf("backward mutation = %v", err)
		}
		if entity.Content() != body || !entity.Tags().Contains(tag) || !entity.Metadata().UpdatedAt().Equal(later) {
			t.Fatal("failed mutation changed the entity")
		}
	}
	if err := entity.Rewrite("", later); err != nil {
		t.Fatal(err)
	}
	if err := entity.ReplaceTags(shared.Tags{}, later); err != nil {
		t.Fatal(err)
	}
	if entity.Content() != "" || !entity.Tags().IsEmpty() {
		t.Fatal("optional content and tags were not cleared")
	}
}

func TestInboxRejectsMissingIdentityAndMetadata(t *testing.T) {
	at := time.Now()
	metadata, err := shared.NewEntityMetadata(at, at)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		id       uuid.UUID
		metadata shared.EntityMetadata
	}{
		{uuid.Nil(), metadata},
		{uuid.New(), shared.EntityMetadata{}},
	} {
		if _, err := inbox.NewInbox(tc.id, "", shared.Tags{}, tc.metadata); !errors.Is(err, inbox.ErrInvalidInbox) {
			t.Fatalf("invalid inbox = %v", err)
		}
	}
	var uninitialized *inbox.Inbox
	if err := uninitialized.Rewrite("content", at); !errors.Is(err, inbox.ErrInvalidInbox) {
		t.Fatalf("nil receiver rewrite = %v", err)
	}
}
