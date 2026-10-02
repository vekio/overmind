package parser

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
)

type inlineMacroKind uint8

const (
	inlineMacroLink inlineMacroKind = iota
	inlineMacroCrossReference
)

type inlineMacroMatch struct {
	kind                   inlineMacroKind
	start, end             int
	targetStart, targetEnd int
	labelStart, labelEnd   int
}

func (p *inlineParser) matchInlineMacro(start, end int) (inlineMacroMatch, bool) {
	if start > 0 {
		previous, _ := utf8.DecodeLastRuneInString(p.source[:start])
		if unicode.IsLetter(previous) || unicode.IsDigit(previous) || previous == '_' {
			return inlineMacroMatch{}, false
		}
	}
	if strings.HasPrefix(p.source[start:end], "<<") {
		closeOffset := strings.Index(p.source[start+2:end], ">>")
		if closeOffset < 0 {
			return inlineMacroMatch{}, false
		}
		closeAt := start + 2 + closeOffset
		targetStart, targetEnd := trimInlineBounds(p.source, start+2, closeAt)
		labelStart, labelEnd := closeAt, closeAt
		if comma := strings.IndexByte(p.source[targetStart:targetEnd], ','); comma >= 0 {
			comma += targetStart
			targetStart, targetEnd = trimInlineBounds(p.source, targetStart, comma)
			labelStart, labelEnd = trimInlineBounds(p.source, comma+1, closeAt)
		}
		if targetStart == targetEnd {
			return inlineMacroMatch{}, false
		}
		return inlineMacroMatch{kind: inlineMacroCrossReference, start: start, end: closeAt + 2, targetStart: targetStart, targetEnd: targetEnd, labelStart: labelStart, labelEnd: labelEnd}, true
	}

	targetOffset, kind, matched := inlineMacroPrefix(p.source[start:end])
	if !matched {
		return inlineMacroMatch{}, false
	}
	targetStart := start + targetOffset
	openOffset := strings.IndexByte(p.source[targetStart:end], '[')
	if openOffset < 0 {
		return inlineMacroMatch{}, false
	}
	openAt := targetStart + openOffset
	closeOffset := strings.IndexByte(p.source[openAt+1:end], ']')
	if closeOffset < 0 || openAt == targetStart {
		return inlineMacroMatch{}, false
	}
	closeAt := openAt + 1 + closeOffset
	return inlineMacroMatch{
		kind: kind, start: start, end: closeAt + 1,
		targetStart: targetStart, targetEnd: openAt,
		labelStart: openAt + 1, labelEnd: closeAt,
	}, true
}

func inlineMacroPrefix(source string) (int, inlineMacroKind, bool) {
	for _, prefix := range []string{"https://", "http://", "ftp://", "irc://", "mailto:"} {
		if strings.HasPrefix(source, prefix) {
			return 0, inlineMacroLink, true
		}
	}
	if strings.HasPrefix(source, "link:") {
		return len("link:"), inlineMacroLink, true
	}
	if strings.HasPrefix(source, "xref:") {
		return len("xref:"), inlineMacroCrossReference, true
	}
	return 0, inlineMacroLink, false
}

func (m inlineMacroMatch) inline(p *inlineParser, children []ast.Inline) ast.Inline {
	source := p.span(m.start, m.end)
	targetSource := p.span(m.targetStart, m.targetEnd)
	labelSource := p.span(m.labelStart, m.labelEnd)
	target := p.source[m.targetStart:m.targetEnd]
	if m.kind == inlineMacroCrossReference {
		return &ast.CrossReference{Source: source, TargetSource: targetSource, LabelSource: labelSource, Target: target, Children: children}
	}
	return &ast.Link{Source: source, TargetSource: targetSource, LabelSource: labelSource, Target: target, Children: children}
}

func trimInlineBounds(source string, start, end int) (int, int) {
	for start < end && (source[start] == ' ' || source[start] == '\t') {
		start++
	}
	for end > start && (source[end-1] == ' ' || source[end-1] == '\t') {
		end--
	}
	return start, end
}
