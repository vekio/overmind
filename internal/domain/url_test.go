package domain_test

import (
	"errors"
	"testing"

	"github.com/vekio/overmind/internal/domain"
)

func TestURLRequiresHTTPOrHTTPS(t *testing.T) {
	if _, err := domain.NewURL("file:///tmp/note"); !errors.Is(err, domain.ErrInvalidURL) {
		t.Fatalf("non-HTTP URL = %v", err)
	}
	url, err := domain.NewURL(" https://example.com/page ")
	if err != nil || url.String() != "https://example.com/page" {
		t.Fatalf("URL = %q, %v", url, err)
	}
}
