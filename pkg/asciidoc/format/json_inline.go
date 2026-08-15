package format

import (
	"fmt"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

func makeJSONInlines(inlines []ast.Inline) ([]jsonInline, error) {
	result := make([]jsonInline, 0, len(inlines))
	for _, inline := range inlines {
		node, err := makeJSONInline(inline)
		if err != nil {
			return nil, err
		}
		result = append(result, node)
	}
	return result, nil
}

func makeJSONInline(inline ast.Inline) (jsonInline, error) {
	switch inline := inline.(type) {
	case *ast.Text:
		value := inline.Value
		return jsonInline{Kind: "text", Source: makeJSONSpan(inline.Source), Value: &value}, nil
	case *ast.Strong:
		return makeJSONFormattedInline("strong", inline.Source, inline.ContentSource, inline.Children)
	case *ast.Emphasis:
		return makeJSONFormattedInline("emphasis", inline.Source, inline.ContentSource, inline.Children)
	case *ast.Monospace:
		return makeJSONFormattedInline("monospace", inline.Source, inline.ContentSource, inline.Children)
	case *ast.Link:
		return makeJSONInlineMacro("link", inline.Source, inline.TargetSource, inline.LabelSource, inline.Target, inline.Children)
	case *ast.CrossReference:
		return makeJSONInlineMacro("cross_reference", inline.Source, inline.TargetSource, inline.LabelSource, inline.Target, inline.Children)
	default:
		return jsonInline{}, fmt.Errorf("encode unsupported AST inline %T", inline)
	}
}

func makeJSONFormattedInline(kind string, source, contentSource ast.Span, children []ast.Inline) (jsonInline, error) {
	encodedChildren, err := makeJSONInlines(children)
	if err != nil {
		return jsonInline{}, err
	}
	encodedContentSource := makeJSONSpan(contentSource)
	return jsonInline{Kind: kind, Source: makeJSONSpan(source), ContentSource: &encodedContentSource, Children: &encodedChildren}, nil
}

func makeJSONInlineMacro(kind string, source, targetSource, labelSource ast.Span, target string, children []ast.Inline) (jsonInline, error) {
	encodedChildren, err := makeJSONInlines(children)
	if err != nil {
		return jsonInline{}, err
	}
	encodedTargetSource := makeJSONSpan(targetSource)
	encodedLabelSource := makeJSONSpan(labelSource)
	return jsonInline{Kind: kind, Source: makeJSONSpan(source), TargetSource: &encodedTargetSource, LabelSource: &encodedLabelSource, Target: &target, Children: &encodedChildren}, nil
}
