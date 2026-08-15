package domain

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidDocumentID = errors.New("invalid document id")

const (
	AttributePrefix    = "overmind-"
	AttributeID        = AttributePrefix + "id"
	AttributeType      = AttributePrefix + "type"
	AttributeTitle     = AttributePrefix + "title"
	AttributeArea      = AttributePrefix + "area"
	AttributeCreatedAt = AttributePrefix + "created-at"
)

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
