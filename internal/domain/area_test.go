package domain

import (
	"errors"
	"testing"
)

func TestNewAreaNormalizesHierarchy(t *testing.T) {
	area, err := NewArea(" Knowledge Base / Go / Testing ")
	if err != nil {
		t.Fatalf("NewArea() error = %v", err)
	}
	if got, want := area.Name(), "knowledge-base"; got != want {
		t.Fatalf("Area.Name() = %q, want %q", got, want)
	}
	if got, want := area.String(), "knowledge-base/go/testing"; got != want {
		t.Fatalf("Area.String() = %q, want %q", got, want)
	}
	if area.IsZero() {
		t.Fatal("Area.IsZero() = true for valid area")
	}
}

func TestAreaSubareaTraversesHierarchy(t *testing.T) {
	area, _ := NewArea("knowledge/go/testing")

	subarea, ok := area.Subarea()
	if !ok || subarea.String() != "go/testing" {
		t.Fatalf("first Subarea() = (%q, %t)", subarea, ok)
	}
	subarea, ok = subarea.Subarea()
	if !ok || subarea.String() != "testing" {
		t.Fatalf("second Subarea() = (%q, %t)", subarea, ok)
	}
	if _, ok := subarea.Subarea(); ok {
		t.Fatal("leaf Subarea() ok = true")
	}
}

func TestAreaEquality(t *testing.T) {
	first, _ := NewArea("Knowledge/Go")
	second, _ := NewArea("knowledge/go")
	other, _ := NewArea("knowledge/Rust")

	if !first.Equal(second) {
		t.Fatal("equivalent areas are not equal")
	}
	if first.Equal(other) {
		t.Fatal("different areas are equal")
	}
}

func TestNewAreaRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"", "   ", "/", "knowledge//go", "knowledge/---"} {
		t.Run(value, func(t *testing.T) {
			if _, err := NewArea(value); !errors.Is(err, ErrInvalidArea) {
				t.Fatalf("NewArea(%q) error = %v, want %v", value, err, ErrInvalidArea)
			}
		})
	}
}

func TestZeroArea(t *testing.T) {
	var area Area
	if !area.IsZero() || area.Name() != "" || area.String() != "" {
		t.Fatalf("zero Area = name %q, value %q, IsZero %t", area.Name(), area.String(), area.IsZero())
	}
	if _, ok := area.Subarea(); ok {
		t.Fatal("zero Area.Subarea() ok = true")
	}
}
