// Package idgenerator generates identifiers used by Overmind documents.
package idgenerator

import (
	"git.casta.me/alberto/overmind/internal/ports"
	"uuid"
)

var _ ports.IDGenerator = Generator{}

// Generator creates random UUIDs.
type Generator struct{}

// New creates an ID generator.
func New() Generator {
	return Generator{}
}

// Generate returns a random UUID version 4.
func (Generator) Generate() uuid.UUID {
	return uuid.NewV4()
}
