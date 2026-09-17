package domain

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidDocumentID = errors.New("invalid document id")

// DocumentID identifies a domain object represented by an AsciiDoc document.
// Its zero value is invalid.
type DocumentID struct {
	value string
}

// NewDocumentID creates a valid document identifier.
func NewDocumentID(value string) (DocumentID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return DocumentID{}, fmt.Errorf("%w: value is required", ErrInvalidDocumentID)
	}
	if strings.ContainsAny(value, "\r\n\t ") {
		return DocumentID{}, fmt.Errorf("%w: value must not contain whitespace", ErrInvalidDocumentID)
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' ||
			character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' ||
			character == '-' || character == '_' {
			continue
		}
		return DocumentID{}, fmt.Errorf("%w: value must contain only letters, numbers, hyphens, or underscores", ErrInvalidDocumentID)
	}
	return DocumentID{value: value}, nil
}

// String returns the identifier value.
func (id DocumentID) String() string { return id.value }

// MarshalText encodes the identifier for JSON, YAML, and text-oriented
// adapters.
func (id DocumentID) MarshalText() ([]byte, error) {
	if id.value == "" {
		return nil, fmt.Errorf("document id is invalid")
	}
	return []byte(id.value), nil
}
