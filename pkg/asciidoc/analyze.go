package asciidoc

import (
	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/semantic"
)

// Analysis is the public semantic result produced by Analyze.
type Analysis = semantic.Analysis

// Analyze derives semantic information from a parsed document without
// mutating it. Use it when an AST already exists; most applications can call
// Process to parse and analyze in one operation.
func Analyze(document *ast.Document) Analysis {
	return semantic.Analyze(document)
}
