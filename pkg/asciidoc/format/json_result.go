package format

import (
	"fmt"

	"git.casta.me/alberto/overmind/pkg/asciidoc"
)

func makeJSONResult(result asciidoc.Result) (jsonResult, error) {
	if result.Document == nil {
		return jsonResult{}, fmt.Errorf("parser returned a nil document")
	}
	diagnostics := make([]jsonDiagnostic, 0, len(result.Diagnostics))
	for _, item := range result.Diagnostics {
		diagnostics = append(diagnostics, jsonDiagnostic{Severity: item.Severity.String(), Message: item.Message, Source: makeJSONSpan(item.Source)})
	}
	blocks, err := makeJSONBlocks(result.Document.Blocks)
	if err != nil {
		return jsonResult{}, err
	}
	document := jsonDocument{Kind: "document", Source: makeJSONSpan(result.Document.Source), Blocks: blocks}
	if result.Document.Title != nil {
		document.Title = &jsonDocumentTitle{Kind: "document_title", Source: makeJSONSpan(result.Document.Title.Source), TitleSource: makeJSONSpan(result.Document.Title.TitleSource), Text: result.Document.Title.Text}
	}
	return jsonResult{Document: document, Diagnostics: diagnostics}, nil
}
