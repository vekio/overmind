package format

import (
	"fmt"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

func makeJSONBlocks(blocks []ast.Block) ([]any, error) {
	result := make([]any, 0, len(blocks))
	for _, block := range blocks {
		node, err := makeJSONNode(block)
		if err != nil {
			return nil, err
		}
		result = append(result, node)
	}
	return result, nil
}

func makeJSONNode(block ast.Block) (any, error) {
	switch block := block.(type) {
	case *ast.Section:
		blocks, err := makeJSONBlocks(block.Blocks)
		if err != nil {
			return nil, err
		}
		return jsonSection{Kind: "section", Source: makeJSONSpan(block.Source), HeadingSource: makeJSONSpan(block.HeadingSource), ContentSource: makeJSONSpan(block.ContentSource), TitleSource: makeJSONSpan(block.TitleSource), Level: block.Level, Title: block.Title, Metadata: makeJSONBlockMetadata(block.Metadata), Blocks: blocks}, nil
	case *ast.Paragraph:
		inlines, err := makeJSONInlines(block.Inlines)
		if err != nil {
			return nil, err
		}
		return jsonParagraph{Kind: "paragraph", Source: makeJSONSpan(block.Source), Text: block.Text, Inlines: inlines}, nil
	case *ast.ThematicBreak:
		return jsonBreak{Kind: "thematic_break", Source: makeJSONSpan(block.Source)}, nil
	case *ast.PageBreak:
		return jsonBreak{Kind: "page_break", Source: makeJSONSpan(block.Source)}, nil
	case *ast.DelimitedBlock:
		result := jsonDelimitedBlock{Kind: "delimited_block", Source: makeJSONSpan(block.Source), OpeningSource: makeJSONSpan(block.OpeningSource), ContentSource: makeJSONSpan(block.ContentSource), BlockKind: block.Kind.String(), ContentModel: block.ContentModel.String(), Marker: block.Marker, Metadata: makeJSONBlockMetadata(block.Metadata), Content: block.Content, Closed: block.Closed}
		if block.Closed {
			closing := makeJSONSpan(block.ClosingSource)
			result.ClosingSource = &closing
		}
		return result, nil
	case *ast.Table:
		return makeJSONTable(block)
	case *ast.List:
		items, err := makeJSONListItems(block.Items)
		if err != nil {
			return nil, err
		}
		return jsonList{Kind: "list", Source: makeJSONSpan(block.Source), ListKind: block.Kind.String(), ListLevel: block.Level, Metadata: makeJSONBlockMetadata(block.Metadata), Items: items}, nil
	case *ast.DescriptionList:
		items, err := makeJSONDescriptionListItems(block.Items)
		if err != nil {
			return nil, err
		}
		return jsonList{Kind: "description_list", Source: makeJSONSpan(block.Source), ListLevel: block.Level, Metadata: makeJSONBlockMetadata(block.Metadata), Items: items}, nil
	case *ast.Admonition:
		inlines, err := makeJSONInlines(block.Inlines)
		if err != nil {
			return nil, err
		}
		return jsonAdmonition{Kind: "admonition", Source: makeJSONSpan(block.Source), ContentSource: makeJSONSpan(block.ContentSource), LabelSource: makeJSONSpan(block.LabelSource), AdmonitionKind: block.Kind.String(), Label: block.Label, Text: block.Text, Inlines: inlines, Metadata: makeJSONBlockMetadata(block.Metadata)}, nil
	case *ast.BlockMacro:
		return jsonBlockMacro{Kind: "block_macro", Source: makeJSONSpan(block.Source), NameSource: makeJSONSpan(block.NameSource), TargetSource: makeJSONSpan(block.TargetSource), AttributesSource: makeJSONSpan(block.AttributesSource), Name: block.Name, Target: block.Target, Attributes: block.Attributes, Metadata: makeJSONBlockMetadata(block.Metadata)}, nil
	case *ast.AttributeEntry:
		result := jsonAttributeEntry{Kind: "attribute_entry", Source: makeJSONSpan(block.Source), NameSource: makeJSONSpan(block.NameSource), Operation: block.Operation.String(), Name: block.Name, Value: block.Value, Header: block.Header}
		if block.Value != "" {
			valueSource := makeJSONSpan(block.ValueSource)
			result.ValueSource = &valueSource
		}
		return result, nil
	default:
		return nil, fmt.Errorf("encode unsupported AST block %T", block)
	}
}
