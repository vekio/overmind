package edit

import (
	"bytes"
	"fmt"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

// ReplaceSectionContent returns an edit that replaces a section's complete
// body. The replacement uses the line ending nearest to the heading and is
// separated from both the heading and a following section by a blank line.
// Nested sections are part of ContentSource and are therefore replaced too.
// The function only prepares a TextEdit; call Apply to produce updated bytes.
func ReplaceSectionContent(source []byte, section *ast.Section, content []byte) (TextEdit, error) {
	if section == nil {
		return TextEdit{}, fmt.Errorf("replace section content: nil section")
	}
	if err := validateSectionSpans(source, section); err != nil {
		return TextEdit{}, err
	}

	ending := sectionLineEnding(source, section)
	normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))
	normalized = bytes.TrimRight(normalized, "\n")
	normalized = bytes.ReplaceAll(normalized, []byte("\n"), ending)

	prefixCount := 1
	if section.ContentSource.Start.Offset == section.HeadingSource.End.Offset {
		prefixCount = 2
	}
	replacement := make([]byte, 0, len(normalized)+(prefixCount+2)*len(ending))
	for range prefixCount {
		replacement = append(replacement, ending...)
	}
	if len(normalized) > 0 {
		replacement = append(replacement, normalized...)
		replacement = append(replacement, ending...)
		replacement = append(replacement, ending...)
	}

	return TextEdit{Source: section.ContentSource, Replacement: replacement}, nil
}

func validateSectionSpans(source []byte, section *ast.Section) error {
	headingStart, headingEnd := section.HeadingSource.Start.Offset, section.HeadingSource.End.Offset
	contentStart, contentEnd := section.ContentSource.Start.Offset, section.ContentSource.End.Offset
	if headingStart < 0 || headingEnd < headingStart || contentStart < headingEnd || contentEnd < contentStart || contentEnd > len(source) {
		return fmt.Errorf("replace section content: section spans are outside the %d-byte source", len(source))
	}
	gap := source[headingEnd:contentStart]
	if len(gap) > 0 && !bytes.Equal(gap, []byte("\n")) && !bytes.Equal(gap, []byte("\r")) && !bytes.Equal(gap, []byte("\r\n")) {
		return fmt.Errorf("replace section content: content does not start after the heading line ending")
	}
	return nil
}

func sectionLineEnding(source []byte, section *ast.Section) []byte {
	headingEnd := section.HeadingSource.End.Offset
	if ending := lineEndingAt(source, headingEnd); ending != nil {
		return ending
	}
	for cursor := section.HeadingSource.Start.Offset - 1; cursor >= 0; cursor-- {
		if source[cursor] == '\n' {
			if cursor > 0 && source[cursor-1] == '\r' {
				return []byte("\r\n")
			}
			return []byte("\n")
		}
		if source[cursor] == '\r' {
			return []byte("\r")
		}
	}
	for cursor := headingEnd; cursor < len(source); cursor++ {
		if ending := lineEndingAt(source, cursor); ending != nil {
			return ending
		}
	}
	return []byte("\n")
}

func lineEndingAt(source []byte, at int) []byte {
	if at < 0 || at >= len(source) {
		return nil
	}
	if source[at] == '\r' {
		if at+1 < len(source) && source[at+1] == '\n' {
			return []byte("\r\n")
		}
		return []byte("\r")
	}
	if source[at] == '\n' {
		return []byte("\n")
	}
	return nil
}
