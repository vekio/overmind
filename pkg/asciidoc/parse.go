package asciidoc

import (
	"bytes"
	"io"

	"github.com/vekio/overmind/pkg/asciidoc/parser"
)

// Result contains the partial or complete AST and all parser diagnostics.
// It aliases parser.Result so callers can use the root package as a facade.
type Result = parser.Result

// Parse parses an AsciiDoc document held in memory without semantic analysis.
// Most applications should use Process; Parse is intended for syntax-only
// consumers and tools that want to control semantic analysis explicitly.
func Parse(source []byte) Result {
	return ParseReader(bytes.NewReader(source))
}

// ParseReader parses an AsciiDoc document from reader without semantic
// analysis. The reader is consumed and not retained.
func ParseReader(reader io.Reader) Result {
	return parser.New(reader).Parse()
}
