// Package id generates identifiers used by Overmind documents.
package id

import "uuid"

// New returns a random UUID version 4.
func New() uuid.UUID {
	return uuid.NewV4()
}
