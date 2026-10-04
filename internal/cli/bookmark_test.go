package cli

import (
	"errors"
	"testing"

	"github.com/vekio/overmind/internal/domain/bookmarks"
)

func TestBookmarkCommandPassesURLAndPrintsID(t *testing.T) {
	client := &otherCommandsClient{}
	got := runCLICommand(t, newBookmarkCommand(fixedClient(client)), []string{"bookmark", "--tag", "Reading", "https://example.com"}, "")
	if got != "https://example.com\nID: 11111111-1111-4111-8111-111111111111\n" || client.bookmarkURL != "https://example.com" {
		t.Fatalf("bookmark output=%q URL=%q", got, client.bookmarkURL)
	}
}

func TestRootBookmarkCommandPersistsDocumentAndProjection(t *testing.T) {
	fixture := newCommandFixture(t)
	url := "https://example.com/a?q=1&lang=es"
	note := fixture.create("bookmark", []string{"bookmark", "--tag", "Reading", "--tag", "Work", url}, "", []string{"reading", "work"})
	var storedURL string
	if err := fixture.db().QueryRow("SELECT url FROM bookmarks WHERE note_id=?", note.ID.String()).Scan(&storedURL); err != nil || storedURL != url {
		t.Fatalf("bookmark projection=%q err=%v", storedURL, err)
	}
	if _, err := fixture.run([]string{"bookmark", "file:///tmp/note"}, ""); !errors.Is(err, bookmarks.ErrInvalidURL) {
		t.Fatalf("invalid URL=%v", err)
	}
	fixture.requireSingleNote()
}
