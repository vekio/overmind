package format

import "github.com/vekio/overmind/pkg/asciidoc/ast"

func makeJSONListItems(items []*ast.ListItem) ([]jsonListItem, error) {
	result := make([]jsonListItem, 0, len(items))
	for _, item := range items {
		inlines, err := makeJSONInlines(item.Inlines)
		if err != nil {
			return nil, err
		}
		blocks, err := makeJSONBlocks(item.Blocks)
		if err != nil {
			return nil, err
		}
		principalSource := makeJSONSpan(item.PrincipalSource)
		principal := item.Principal
		result = append(result, jsonListItem{Kind: "list_item", Source: makeJSONSpan(item.Source), MarkerSource: makeJSONSpan(item.MarkerSource), PrincipalSource: &principalSource, Marker: item.Marker, Principal: &principal, Inlines: &inlines, Blocks: blocks})
	}
	return result, nil
}

func makeJSONDescriptionListItems(items []*ast.DescriptionListItem) ([]jsonListItem, error) {
	result := make([]jsonListItem, 0, len(items))
	for _, item := range items {
		termInlines, err := makeJSONInlines(item.TermInlines)
		if err != nil {
			return nil, err
		}
		descriptionInlines, err := makeJSONInlines(item.DescriptionInlines)
		if err != nil {
			return nil, err
		}
		blocks, err := makeJSONBlocks(item.Blocks)
		if err != nil {
			return nil, err
		}
		termSource := makeJSONSpan(item.TermSource)
		descriptionSource := makeJSONSpan(item.DescriptionSource)
		term, description := item.Term, item.Description
		result = append(result, jsonListItem{Kind: "description_list_item", Source: makeJSONSpan(item.Source), MarkerSource: makeJSONSpan(item.MarkerSource), TermSource: &termSource, DescriptionSource: &descriptionSource, Marker: item.Marker, Term: &term, Description: &description, TermInlines: &termInlines, DescriptionInlines: &descriptionInlines, Blocks: blocks})
	}
	return result, nil
}
