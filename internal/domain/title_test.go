package domain

import (
	"errors"
	"testing"
)

func TestNewTitleNormalizesValueAndGeneratesSlug(t *testing.T) {
	title, err := NewTitle("  Mi primera página  ")
	if err != nil {
		t.Fatalf("NewTitle() error = %v", err)
	}
	if got, want := title.String(), "Mi primera página"; got != want {
		t.Fatalf("Title.String() = %q, want %q", got, want)
	}
	if got, want := title.Slug(), "mi-primera-pagina"; got != want {
		t.Fatalf("Title.Slug() = %q, want %q", got, want)
	}
	if title.IsZero() {
		t.Fatal("Title.IsZero() = true for valid title")
	}
}

func TestNewTitleRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"", "   ", "---", "Title\nattribute", "Title\rattribute"} {
		t.Run(value, func(t *testing.T) {
			if _, err := NewTitle(value); !errors.Is(err, ErrInvalidTitle) {
				t.Fatalf("NewTitle(%q) error = %v, want %v", value, err, ErrInvalidTitle)
			}
		})
	}
}

func TestZeroTitle(t *testing.T) {
	var title Title
	if !title.IsZero() || title.String() != "" || title.Slug() != "" {
		t.Fatalf("zero Title = value %q, slug %q, IsZero %t", title.String(), title.Slug(), title.IsZero())
	}
}
