package domain_test

import (
	"errors"
	"testing"

	"github.com/vekio/overmind/internal/domain"
)

func TestAreaNormalizesHierarchyAndExposesSubarea(t *testing.T) {
	area, err := domain.NewArea("  Proyectos / Diseño / Ideas  ")
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
	if _, err := domain.NewArea("work//ideas"); !errors.Is(err, domain.ErrInvalidArea) {
		t.Fatalf("empty area segment = %v", err)
	}
}
