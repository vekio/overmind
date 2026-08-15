// Package diagnostic defines parser and semantic messages associated with
// source spans.
package diagnostic

import (
	"fmt"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

// Severity identifies the impact of a diagnostic.
type Severity uint8

const (
	SeverityUnknown Severity = iota
	SeverityWarning
	SeverityError
)

// String returns a stable lowercase severity name.
func (s Severity) String() string {
	switch s {
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

// Diagnostic describes a parser problem at a source range.
type Diagnostic struct {
	Severity Severity
	Message  string
	Source   ast.Span
}

// String returns a human-readable diagnostic.
func (d Diagnostic) String() string {
	return fmt.Sprintf("%s: %s at %s", d.Severity, d.Message, d.Source)
}
