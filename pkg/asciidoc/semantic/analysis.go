package semantic

import (
	"sort"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
)

// Analysis contains semantic information derived from a parsed document.
// Document is the original syntax tree and is not mutated.
type Analysis struct {
	// Document is the input AST and is not mutated by analysis.
	Document *ast.Document
	// Header contains the title and effective header attributes.
	Header Header
	// Anchors indexes explicit anchors attached to supported AST nodes.
	Anchors AnchorIndex
	// References contains every parsed cross-reference in source order.
	References []Reference
	// Diagnostics contains semantic diagnostics only.
	Diagnostics []diagnostic.Diagnostic
}

// Header contains the effective semantic state of the document header.
type Header struct {
	// Title aliases Document.Title and may be nil.
	Title *ast.DocumentTitle
	// Attributes contains effective header attributes and their histories.
	Attributes AttributeSet
}

// Analyze derives semantic information from document without mutating it.
// A nil document produces an empty analysis.
func Analyze(document *ast.Document) Analysis {
	analysis := Analysis{
		Document:    document,
		References:  make([]Reference, 0),
		Diagnostics: make([]diagnostic.Diagnostic, 0),
	}
	if document == nil {
		analysis.Header.Attributes = analyzeHeaderAttributes(nil)
		analysis.Anchors, _ = analyzeAnchors(nil)
		return analysis
	}
	analysis.Header = Header{
		Title:      document.Title,
		Attributes: analyzeHeaderAttributes(document),
	}
	var anchorDiagnostics []diagnostic.Diagnostic
	analysis.Anchors, anchorDiagnostics = analyzeAnchors(document)
	analysis.Diagnostics = append(analysis.Diagnostics, anchorDiagnostics...)
	var referenceDiagnostics []diagnostic.Diagnostic
	analysis.References, referenceDiagnostics = analyzeReferences(document, analysis.Anchors)
	analysis.Diagnostics = append(analysis.Diagnostics, referenceDiagnostics...)
	sort.SliceStable(analysis.Diagnostics, func(left, right int) bool {
		return analysis.Diagnostics[left].Source.Start.Offset < analysis.Diagnostics[right].Source.Start.Offset
	})
	return analysis
}

// HasErrors reports whether semantic analysis produced an error diagnostic.
func (a Analysis) HasErrors() bool {
	for _, item := range a.Diagnostics {
		if item.Severity == diagnostic.SeverityError {
			return true
		}
	}
	return false
}
