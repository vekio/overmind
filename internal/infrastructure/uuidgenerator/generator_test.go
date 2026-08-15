package uuidgenerator

import (
	"testing"

	"github.com/google/uuid"
)

func TestGeneratorReturnsUUIDV4(t *testing.T) {
	generated, err := New().Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	id, err := uuid.Parse(generated)
	if err != nil {
		t.Fatalf("uuid.Parse(%q) error = %v", generated, err)
	}
	if got, want := id.Version(), uuid.Version(4); got != want {
		t.Fatalf("UUID version = %v, want %v", got, want)
	}
	if got, want := id.Variant(), uuid.RFC4122; got != want {
		t.Fatalf("UUID variant = %v, want %v", got, want)
	}
}
