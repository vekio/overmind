package domain

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrInvalidURL indicates an invalid bookmark URL.
var ErrInvalidURL = errors.New("invalid URL")

// URL represents a normalized HTTP or HTTPS URL.
type URL struct {
	value string
}

// NewURL creates a validated HTTP or HTTPS URL.
func NewURL(value string) (URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return URL{}, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if parsed.Host == "" || parsed.Scheme != "http" && parsed.Scheme != "https" {
		return URL{}, fmt.Errorf("%w: expected an HTTP or HTTPS URL", ErrInvalidURL)
	}

	return URL{value: parsed.String()}, nil
}

// IsZero reports whether the URL is uninitialized.
func (value URL) IsZero() bool {
	return value.value == ""
}

// String returns the normalized URL.
func (value URL) String() string {
	return value.value
}
