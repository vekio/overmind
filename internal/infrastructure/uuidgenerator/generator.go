// Package uuidgenerator implements identifier generation using UUID v7.
package uuidgenerator

import (
	"github.com/google/uuid"

	"git.casta.me/alberto/overmind/internal/ports"
)

var _ ports.IDGenerator = Generator{}

// Generator creates time-ordered UUID v7 identifiers.
type Generator struct{}

// New creates a UUID generator.
func New() Generator { return Generator{} }

// Generate returns a new UUID v7 identifier.
func (Generator) Generate() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
