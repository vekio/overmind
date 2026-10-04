package bookmarks_test

import (
	"errors"
	"testing"

	"github.com/vekio/overmind/internal/domain/bookmarks"
)

func TestURLRequiresHTTPOrHTTPS(t *testing.T) {
	if _, err := bookmarks.NewURL("file:///tmp/note"); !errors.Is(err, bookmarks.ErrInvalidURL) {
		t.Fatalf("non-HTTP URL = %v", err)
	}
	url, err := bookmarks.NewURL(" https://example.com/page ")
	if err != nil || url.String() != "https://example.com/page" {
		t.Fatalf("URL = %q, %v", url, err)
	}
}
