package format

import "github.com/vekio/overmind/pkg/asciidoc/ast"

func makeJSONBlockMetadata(metadata ast.BlockMetadata) jsonBlockMetadata {
	result := jsonBlockMetadata{AttributeLists: make([]jsonAttributeList, 0, len(metadata.AttributeLists))}
	if metadata.Title != nil || metadata.Anchor != nil || len(metadata.AttributeLists) > 0 {
		source := makeJSONSpan(metadata.Source)
		result.Source = &source
	}
	if metadata.Title != nil {
		result.Title = &jsonBlockTitle{Source: makeJSONSpan(metadata.Title.Source), TitleSource: makeJSONSpan(metadata.Title.TitleSource), Text: metadata.Title.Text}
	}
	if metadata.Anchor != nil {
		anchor := &jsonAnchor{Source: makeJSONSpan(metadata.Anchor.Source), IDSource: makeJSONSpan(metadata.Anchor.IDSource), ID: metadata.Anchor.ID, RefText: metadata.Anchor.RefText}
		if metadata.Anchor.RefText != "" {
			refTextSource := makeJSONSpan(metadata.Anchor.RefTextSource)
			anchor.RefTextSource = &refTextSource
		}
		result.Anchor = anchor
	}
	for _, attributeList := range metadata.AttributeLists {
		entries := make([]jsonAttribute, 0, len(attributeList.Entries))
		for _, entry := range attributeList.Entries {
			entries = append(entries, jsonAttribute{Source: makeJSONSpan(entry.Source), Value: entry.Value})
		}
		result.AttributeLists = append(result.AttributeLists, jsonAttributeList{Source: makeJSONSpan(attributeList.Source), Entries: entries})
	}
	return result
}
