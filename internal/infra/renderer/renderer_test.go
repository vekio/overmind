package renderer_test

import (
	"context"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain"
	"github.com/vekio/overmind/internal/infra/asciidocnote"
	"github.com/vekio/overmind/internal/infra/renderer"
)

func TestEmptyInboxTemplateCanBeParsedForEditing(t *testing.T) {
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	inbox, err := domain.NewInbox(id, "", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := renderer.New()
	if err != nil {
		t.Fatal(err)
	}
	source, err := renderer.Render(context.Background(), inbox)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(string(source), "= ") {
		t.Fatal("empty capture should not have a generated title")
	}
	record, err := (asciidocnote.Parser{}).Parse(source)
	if err != nil {
		t.Fatalf("parse empty inbox template: %v", err)
	}
	if record.ID != id || record.Kind != domain.NoteKindInbox {
		t.Fatalf("unexpected empty inbox metadata: ID=%s kind=%s", record.ID, record.Kind)
	}
}

func TestBookmarkTemplateCanBeParsedWithoutTitle(t *testing.T) {
	id := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	url, err := domain.NewURL("https://example.com/notes")
	if err != nil {
		t.Fatal(err)
	}
	bookmark, err := domain.NewBookmark(id, url, domain.Tags{}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := renderer.New()
	if err != nil {
		t.Fatal(err)
	}
	source, err := renderer.Render(context.Background(), bookmark)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(string(source), "= ") {
		t.Fatal("bookmark should not have a generated title")
	}
	record, err := (asciidocnote.Parser{}).Parse(source)
	if err != nil {
		t.Fatalf("parse bookmark template: %v", err)
	}
	if record.ID != id || record.Kind != domain.NoteKindBookmark {
		t.Fatalf("unexpected bookmark metadata: ID=%s kind=%s", record.ID, record.Kind)
	}
}
