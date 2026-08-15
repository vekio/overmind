// Package asciidoc is the public facade for processing AsciiDoc documents.
//
// Most applications should call Process or ProcessReader. The returned
// ProcessResult exposes the AST, semantic analysis, and combined diagnostics.
// Parse and ParseReader are lower-level entry points for syntax-only tools;
// Analyze is useful when an application already has an AST.
//
// Source positions always refer to the exact byte slice or stream that was
// processed. After applying source edits, clients must discard the old AST and
// analysis and process the updated bytes again.
package asciidoc
