package edit

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/vekio/overmind/pkg/asciidoc"
	"github.com/vekio/overmind/pkg/asciidoc/ast"
)

func setHeaderAttributeEdit(source []byte, processed asciidoc.ProcessResult, name, value string) (TextEdit, error) {
	if processed.Document == nil {
		return TextEdit{}, fmt.Errorf("set header attribute: nil document")
	}
	if processed.HasErrors() {
		return TextEdit{}, fmt.Errorf("set header attribute: document has errors")
	}
	if !validAttributeName(name) {
		return TextEdit{}, fmt.Errorf("set header attribute: invalid name %q", name)
	}
	if strings.ContainsAny(value, "\r\n") {
		return TextEdit{}, fmt.Errorf("set header attribute %q: value must fit on one line", name)
	}

	declaration := attributeDeclaration(name, value)
	history := processed.Analysis.Header.Attributes.History(name)
	if len(history) != 0 {
		span := history[len(history)-1].Source
		if err := validateSpan(source, span); err != nil {
			return TextEdit{}, fmt.Errorf("set header attribute %q: %w", name, err)
		}
		return TextEdit{Source: span, Replacement: declaration}, nil
	}

	return insertHeaderAttribute(source, processed.Document, declaration), nil
}

func insertHeaderAttribute(source []byte, document *ast.Document, declaration []byte) TextEdit {
	ending := preferredLineEnding(source)
	insertionOffset := 0
	for _, block := range document.Blocks {
		attribute, ok := block.(*ast.AttributeEntry)
		if !ok || attribute == nil || !attribute.Header {
			continue
		}
		insertionOffset = offsetAfterLine(source, attribute.Source.End.Offset)
	}
	if insertionOffset == 0 && document.Title != nil {
		insertionOffset = offsetAfterLine(source, document.Title.Source.End.Offset)
	}

	replacement := make([]byte, 0, len(declaration)+2*len(ending))
	if insertionOffset > 0 && !endsAfterLineEnding(source, insertionOffset) {
		replacement = append(replacement, ending...)
	}
	replacement = append(replacement, declaration...)
	if insertionOffset < len(source) || len(source) == 0 {
		replacement = append(replacement, ending...)
	}
	position := ast.Position{Offset: insertionOffset}
	return TextEdit{Source: ast.Span{Start: position, End: position}, Replacement: replacement}
}

func attributeDeclaration(name, value string) []byte {
	if value == "" {
		return []byte(":" + name + ":")
	}
	return []byte(":" + name + ": " + value)
}

func validAttributeName(name string) bool {
	if name == "" || !attributeWordByte(name[0]) {
		return false
	}
	for index := 1; index < len(name); index++ {
		if !attributeWordByte(name[index]) && name[index] != '-' {
			return false
		}
	}
	return true
}

func attributeWordByte(value byte) bool {
	return value >= 'a' && value <= 'z' ||
		value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9' ||
		value == '_'
}

func validateSpan(source []byte, span ast.Span) error {
	if span.Start.Offset < 0 || span.End.Offset < span.Start.Offset || span.End.Offset > len(source) {
		return fmt.Errorf("attribute span is outside the %d-byte source", len(source))
	}
	return nil
}

func offsetAfterLine(source []byte, offset int) int {
	if offset < 0 || offset > len(source) {
		return 0
	}
	if offset < len(source) && source[offset] == '\r' {
		offset++
		if offset < len(source) && source[offset] == '\n' {
			offset++
		}
		return offset
	}
	if offset < len(source) && source[offset] == '\n' {
		return offset + 1
	}
	return offset
}

func endsAfterLineEnding(source []byte, offset int) bool {
	return offset > 0 && offset <= len(source) && (source[offset-1] == '\n' || source[offset-1] == '\r')
}

func preferredLineEnding(source []byte) []byte {
	if index := bytes.IndexByte(source, '\n'); index >= 0 {
		if index > 0 && source[index-1] == '\r' {
			return []byte("\r\n")
		}
		return []byte("\n")
	}
	if bytes.IndexByte(source, '\r') >= 0 {
		return []byte("\r")
	}
	return []byte("\n")
}
