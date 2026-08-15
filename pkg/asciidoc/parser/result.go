package parser

import (
	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/diagnostic"
)

// Result contains the partial or complete AST and every diagnostic produced
// while parsing. Document is always non-nil.
type Result struct {
	Document    *ast.Document
	Diagnostics []diagnostic.Diagnostic
}

// HasErrors reports whether at least one error diagnostic was produced.
func (r Result) HasErrors() bool {
	for _, item := range r.Diagnostics {
		if item.Severity == diagnostic.SeverityError {
			return true
		}
	}
	return false
}
