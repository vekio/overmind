package domain

import (
	"errors"
	"testing"
)

func TestNewTagNormalizesValue(t *testing.T) {
	tag, err := NewTag(" Diseño de Dominio ")
	if err != nil {
		t.Fatalf("NewTag() error = %v", err)
	}
	if tag.String() != "diseno-de-dominio" {
		t.Fatalf("Tag.String() = %q, want %q", tag, "diseno-de-dominio")
	}
}

func TestNewTagRejectsEmptyNormalizedValue(t *testing.T) {
	if _, err := NewTag("---"); !errors.Is(err, ErrInvalidTag) {
		t.Fatalf("NewTag() error = %v, want %v", err, ErrInvalidTag)
	}
}

func TestTagZeroAndEquality(t *testing.T) {
	first, _ := NewTag("Go Lang")
	second, _ := NewTag("go-lang")
	if !first.Equal(second) {
		t.Fatalf("%q should equal %q", first, second)
	}
	if first.IsZero() || !(Tag{}).IsZero() {
		t.Fatal("Tag.IsZero() returned an unexpected value")
	}
}
