package domain

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidDocumentKind = errors.New("invalid document kind")

// DocumentKind identifies one of the document types supported by Overmind.
type DocumentKind string

const (
	DocumentKindPage DocumentKind = "page"
)

// NewDocumentKind parses a supported document kind.
func NewDocumentKind(value string) (DocumentKind, error) {
	kind := DocumentKind(strings.TrimSpace(value))
	if !kind.IsValid() {
		return "", fmt.Errorf("%w: %q", ErrInvalidDocumentKind, value)
	}
	return kind, nil
}

// IsValid reports whether the kind is supported by the domain model.
func (kind DocumentKind) IsValid() bool {
	switch kind {
	case DocumentKindPage:
		return true
	default:
		return false
	}
}

// String returns the serialized kind value.
func (kind DocumentKind) String() string { return string(kind) }
