package parser

import (
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
	"github.com/vekio/overmind/pkg/asciidoc/lexer"
)

// Parser coordinates the line lexer and construction-specific parsers. A
// Parser consumes its reader and is intended for one call to Parse.
type Parser struct {
	scanner     *lexer.Scanner
	current     lexer.LineToken
	position    ast.Position
	diagnostics []diagnostic.Diagnostic
	parsed      bool
}

// New creates a parser that reads from reader.
func New(reader io.Reader) *Parser {
	return &Parser{
		scanner:  lexer.New(reader),
		position: sourceOrigin(),
	}
}

// Parse consumes the input and returns its partial or complete syntax tree and
// diagnostics. Calling Parse more than once produces an error result.
func (p *Parser) Parse() Result {
	if p == nil {
		parser := New(nil)
		return parser.Parse()
	}
	if p.parsed {
		document := &ast.Document{Source: ast.Span{Start: sourceOrigin(), End: p.position}}
		return Result{
			Document: document,
			Diagnostics: []diagnostic.Diagnostic{{
				Severity: diagnostic.SeverityError,
				Message:  "parser cannot be reused",
				Source:   ast.Span{Start: p.position, End: p.position},
			}},
		}
	}
	p.parsed = true
	p.advance()
	document := p.parseDocument()
	return Result{Document: document, Diagnostics: p.diagnostics}
}

func (p *Parser) advance() {
	if p.scanner == nil {
		p.scanner = lexer.New(nil)
	}
	token, err := p.scanner.Next()
	if err != nil {
		at := ast.Span{Start: p.position, End: p.position}
		p.diagnostics = append(p.diagnostics, diagnostic.Diagnostic{
			Severity: diagnostic.SeverityError,
			Message:  fmt.Sprintf("read AsciiDoc source: %v", err),
			Source:   at,
		})
		p.current = lexer.LineToken{
			Kind: lexer.LineEOF,
			Source: lexer.Span{
				Start: lexerPosition(p.position),
				End:   lexerPosition(p.position),
			},
		}
		return
	}
	p.current = token
	p.position = tokenEndPosition(token)
}

func (p *Parser) appendDiagnostic(severity diagnostic.Severity, message string, source ast.Span) {
	p.diagnostics = append(p.diagnostics, diagnostic.Diagnostic{
		Severity: severity,
		Message:  message,
		Source:   source,
	})
}

func sourceOrigin() ast.Position {
	return ast.Position{Line: 1, Column: 1}
}

func tokenSpan(token lexer.LineToken) ast.Span {
	return ast.Span{Start: position(token.Source.Start), End: position(token.Source.End)}
}

func tokenFragmentSpan(token lexer.LineToken, byteOffset, byteLength int) ast.Span {
	start := ast.Position{
		Offset: token.Source.Start.Offset + byteOffset,
		Line:   token.Source.Start.Line,
		Column: token.Source.Start.Column + utf8.RuneCountInString(token.Raw[:byteOffset]),
	}
	end := ast.Position{
		Offset: start.Offset + byteLength,
		Line:   start.Line,
		Column: start.Column + utf8.RuneCountInString(token.Raw[byteOffset:byteOffset+byteLength]),
	}
	return ast.Span{Start: start, End: end}
}

func tokenEndPosition(token lexer.LineToken) ast.Position {
	if token.Kind == lexer.LineEOF || token.Ending == lexer.LineEndingNone {
		return position(token.Source.End)
	}
	return ast.Position{
		Offset: token.Source.End.Offset + len(token.Ending.Text()),
		Line:   token.Source.End.Line + 1,
		Column: 1,
	}
}

func position(value lexer.Position) ast.Position {
	return ast.Position{Offset: value.Offset, Line: value.Line, Column: value.Column}
}

func lexerPosition(value ast.Position) lexer.Position {
	return lexer.Position{Offset: value.Offset, Line: value.Line, Column: value.Column}
}
