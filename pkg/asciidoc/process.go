package asciidoc

import (
	"bytes"
	"io"
	"sort"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
)

// ProcessResult contains the syntax tree, semantic analysis, and diagnostics
// produced by the complete public processing pipeline. Document is the same
// pointer as Analysis.Document and is always non-nil for parser-produced
// results.
type ProcessResult struct {
	// Document is the partial or complete syntax tree. It aliases
	// Analysis.Document for convenient direct access.
	Document *ast.Document
	// Analysis contains effective header attributes, explicit anchors,
	// resolved references, and semantic diagnostics.
	Analysis Analysis
	// ParseDiagnostics contains syntax and input diagnostics only.
	ParseDiagnostics []diagnostic.Diagnostic
	// Diagnostics combines parser and semantic diagnostics in source order.
	// Warnings are included and do not make HasErrors return true.
	Diagnostics []diagnostic.Diagnostic
}

// HasErrors reports whether parsing or semantic analysis produced an error.
func (r ProcessResult) HasErrors() bool {
	for _, item := range r.Diagnostics {
		if item.Severity == diagnostic.SeverityError {
			return true
		}
	}
	return false
}

// Process parses and semantically analyzes an AsciiDoc document held in
// memory. Semantic analysis runs against the partial AST even when parsing
// reports errors.
func Process(source []byte) ProcessResult {
	return ProcessReader(bytes.NewReader(source))
}

// ProcessReader parses and semantically analyzes AsciiDoc read from reader.
// The reader is consumed and not retained. Clients intending to edit the
// document should keep the original bytes and normally use Process instead.
func ProcessReader(reader io.Reader) ProcessResult {
	parsed := ParseReader(reader)
	analysis := Analyze(parsed.Document)
	diagnostics := make([]diagnostic.Diagnostic, 0, len(parsed.Diagnostics)+len(analysis.Diagnostics))
	diagnostics = append(diagnostics, parsed.Diagnostics...)
	diagnostics = append(diagnostics, analysis.Diagnostics...)
	sort.SliceStable(diagnostics, func(left, right int) bool {
		return diagnostics[left].Source.Start.Offset < diagnostics[right].Source.Start.Offset
	})
	return ProcessResult{
		Document:         parsed.Document,
		Analysis:         analysis,
		ParseDiagnostics: append([]diagnostic.Diagnostic(nil), parsed.Diagnostics...),
		Diagnostics:      diagnostics,
	}
}
