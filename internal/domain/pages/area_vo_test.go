package pages_test

import (
	"errors"
	"testing"

	"github.com/vekio/overmind/internal/domain/pages"
)

func TestAreaNormalizesHierarchyAndExposesSubarea(t *testing.T) {
	area, err := pages.NewArea("  Proyectos / Diseño / Ideas  ")
	if err != nil {
		t.Fatal(err)
	}
	if area.String() != "proyectos/diseno/ideas" || area.Name() != "proyectos" {
		t.Fatalf("area = %q, name = %q", area, area.Name())
	}
	subarea, ok := area.Subarea()
	if !ok || subarea.String() != "diseno/ideas" {
		t.Fatalf("subarea = %q, %t", subarea, ok)
	}
	if _, err := pages.NewArea("work//ideas"); !errors.Is(err, pages.ErrInvalidArea) {
		t.Fatalf("empty area segment = %v", err)
	}
}
