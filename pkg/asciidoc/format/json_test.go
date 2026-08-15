package format_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc"
	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/diagnostic"
	"git.casta.me/alberto/overmind/pkg/asciidoc/format"
)

func TestWriteJSONSerializesCompleteResult(t *testing.T) {
	paragraph := &ast.Paragraph{Source: span(11, 3, 1, 18, 3, 8), Text: "<intro>"}
	thematicBreak := &ast.ThematicBreak{Source: span(40, 7, 1, 43, 7, 4)}
	pageBreak := &ast.PageBreak{Source: span(45, 9, 1, 48, 9, 4)}
	section := &ast.Section{
		Source:      span(20, 5, 1, 48, 9, 4),
		TitleSource: span(23, 5, 4, 35, 5, 16),
		Level:       1,
		Title:       "Installation",
		Blocks:      []ast.Block{thematicBreak, pageBreak},
	}
	result := asciidoc.Result{
		Document: &ast.Document{
			Source: span(0, 1, 1, 48, 9, 4),
			Title: &ast.DocumentTitle{
				Source:      span(0, 1, 1, 9, 1, 10),
				TitleSource: span(2, 1, 3, 9, 1, 10),
				Text:        "Example",
			},
			Blocks: []ast.Block{paragraph, section},
		},
		Diagnostics: []diagnostic.Diagnostic{{
			Severity: diagnostic.SeverityWarning,
			Message:  "example warning",
			Source:   span(20, 5, 1, 22, 5, 3),
		}},
	}

	var output bytes.Buffer
	if err := format.WriteJSON(&output, result); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	if !strings.HasPrefix(output.String(), "{\n  \"document\"") || !strings.HasSuffix(output.String(), "\n") {
		t.Fatalf("JSON is not indented or newline terminated:\n%s", output.String())
	}
	if strings.Contains(output.String(), `\u003c`) || !strings.Contains(output.String(), `"text": "<intro>"`) {
		t.Fatalf("HTML characters were escaped unexpectedly:\n%s", output.String())
	}

	var decoded wireResult
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v; output:\n%s", err, output.String())
	}
	if decoded.Document.Kind != "document" || decoded.Document.Title == nil || decoded.Document.Title.Kind != "document_title" || decoded.Document.Title.Text != "Example" {
		t.Fatalf("document = %+v", decoded.Document)
	}
	if decoded.Document.Source.End != (wirePosition{Offset: 48, Line: 9, Column: 4}) {
		t.Fatalf("document end = %+v", decoded.Document.Source.End)
	}
	if len(decoded.Document.Blocks) != 2 || decoded.Document.Blocks[0].Kind != "paragraph" || decoded.Document.Blocks[0].Text == nil || *decoded.Document.Blocks[0].Text != "<intro>" {
		t.Fatalf("root blocks = %+v", decoded.Document.Blocks)
	}
	sectionNode := decoded.Document.Blocks[1]
	if sectionNode.Kind != "section" || sectionNode.Level == nil || *sectionNode.Level != 1 || sectionNode.Title == nil || *sectionNode.Title != "Installation" {
		t.Fatalf("section = %+v", sectionNode)
	}
	if sectionNode.TitleSource == nil || sectionNode.TitleSource.Start.Offset != 23 || sectionNode.Blocks == nil || len(*sectionNode.Blocks) != 2 {
		t.Fatalf("section metadata or children = %+v", sectionNode)
	}
	if (*sectionNode.Blocks)[0].Kind != "thematic_break" || (*sectionNode.Blocks)[1].Kind != "page_break" {
		t.Fatalf("break nodes = %+v", *sectionNode.Blocks)
	}
	if len(decoded.Diagnostics) != 1 || decoded.Diagnostics[0].Severity != "warning" || decoded.Diagnostics[0].Message != "example warning" || decoded.Diagnostics[0].Source.Start.Offset != 20 {
		t.Fatalf("diagnostics = %+v", decoded.Diagnostics)
	}
}

func TestWriteJSONUsesNullTitleAndEmptyArrays(t *testing.T) {
	result := asciidoc.Result{Document: &ast.Document{Source: span(0, 1, 1, 0, 1, 1)}}
	var output bytes.Buffer
	if err := format.WriteJSON(&output, result); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	if !strings.Contains(output.String(), `"title": null`) || !strings.Contains(output.String(), `"blocks": []`) || !strings.Contains(output.String(), `"diagnostics": []`) {
		t.Fatalf("empty result JSON =\n%s", output.String())
	}
}

func TestWriteJSONSerializesNestedInlines(t *testing.T) {
	paragraph := &ast.Paragraph{
		Source: span(0, 1, 1, 11, 1, 12),
		Text:   "Use *this*",
		Inlines: []ast.Inline{
			&ast.Text{Source: span(0, 1, 1, 4, 1, 5), Value: "Use "},
			&ast.Strong{
				Source:        span(4, 1, 5, 10, 1, 11),
				ContentSource: span(5, 1, 6, 9, 1, 10),
				Children: []ast.Inline{
					&ast.Emphasis{
						Source:        span(5, 1, 6, 9, 1, 10),
						ContentSource: span(6, 1, 7, 8, 1, 9),
						Children: []ast.Inline{
							&ast.Text{Source: span(6, 1, 7, 8, 1, 9), Value: "hi"},
						},
					},
				},
			},
		},
	}
	result := asciidoc.Result{Document: &ast.Document{
		Source: span(0, 1, 1, 11, 1, 12),
		Blocks: []ast.Block{paragraph},
	}}

	var output bytes.Buffer
	if err := format.WriteJSON(&output, result); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	var decoded wireResult
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	node := decoded.Document.Blocks[0]
	if node.Inlines == nil || len(*node.Inlines) != 2 {
		t.Fatalf("paragraph inlines = %+v", node.Inlines)
	}
	strong := (*node.Inlines)[1]
	if strong.Kind != "strong" || strong.ContentSource == nil || strong.Children == nil || len(*strong.Children) != 1 {
		t.Fatalf("strong = %+v", strong)
	}
	emphasis := (*strong.Children)[0]
	if emphasis.Kind != "emphasis" || emphasis.Children == nil || len(*emphasis.Children) != 1 {
		t.Fatalf("emphasis = %+v", emphasis)
	}
	text := (*emphasis.Children)[0]
	if text.Kind != "text" || text.Value == nil || *text.Value != "hi" {
		t.Fatalf("text = %+v", text)
	}
}

func TestWriteJSONSerializesDelimitedBlockMetadata(t *testing.T) {
	block := &ast.DelimitedBlock{
		Source:        span(0, 1, 1, 52, 5, 5),
		OpeningSource: span(26, 3, 1, 30, 3, 5),
		ContentSource: span(31, 4, 1, 47, 5, 1),
		ClosingSource: span(48, 5, 1, 52, 5, 5),
		Kind:          ast.DelimitedBlockListing,
		ContentModel:  ast.ContentModelVerbatim,
		Marker:        "----",
		Metadata: ast.BlockMetadata{
			Source: span(0, 1, 1, 25, 2, 12),
			Title: &ast.BlockTitle{
				Source:      span(0, 1, 1, 8, 1, 9),
				TitleSource: span(1, 1, 2, 8, 1, 9),
				Text:        "Example",
			},
			Anchor: &ast.Anchor{
				Source:        span(9, 2, 1, 25, 2, 17),
				IDSource:      span(11, 2, 3, 13, 2, 5),
				RefTextSource: span(14, 2, 6, 23, 2, 15),
				ID:            "id",
				RefText:       "Reference",
			},
			AttributeLists: []ast.AttributeList{{
				Source: span(14, 2, 6, 25, 2, 17),
				Entries: []ast.Attribute{{
					Source: span(15, 2, 7, 21, 2, 13),
					Value:  "source",
				}},
			}},
		},
		Content: "package main\n",
		Closed:  true,
	}
	result := asciidoc.Result{
		Document: &ast.Document{
			Source: span(0, 1, 1, 52, 5, 5),
			Blocks: []ast.Block{block},
		},
	}

	var output bytes.Buffer
	if err := format.WriteJSON(&output, result); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	var decoded wireResult
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	node := decoded.Document.Blocks[0]
	if node.Kind != "delimited_block" || node.BlockKind == nil || *node.BlockKind != "listing" || node.ContentModel == nil || *node.ContentModel != "verbatim" {
		t.Fatalf("delimited block identity = %+v", node)
	}
	if node.Marker == nil || *node.Marker != "----" || node.Content == nil || *node.Content != "package main\n" || node.Closed == nil || !*node.Closed {
		t.Fatalf("delimited block content = %+v", node)
	}
	if node.OpeningSource == nil || node.OpeningSource.Start.Offset != 26 || node.ContentSource == nil || node.ContentSource.Start.Offset != 31 || node.ClosingSource == nil || node.ClosingSource.Start.Offset != 48 {
		t.Fatalf("delimited block sources = %+v", node)
	}
	if node.Metadata == nil || node.Metadata.Source == nil || node.Metadata.Title == nil || node.Metadata.Title.Text != "Example" || node.Metadata.Anchor == nil || node.Metadata.Anchor.ID != "id" {
		t.Fatalf("metadata = %+v", node.Metadata)
	}
	if node.Metadata.Anchor.RefTextSource == nil || node.Metadata.Anchor.RefText != "Reference" || len(node.Metadata.AttributeLists) != 1 || len(node.Metadata.AttributeLists[0].Entries) != 1 || node.Metadata.AttributeLists[0].Entries[0].Value != "source" {
		t.Fatalf("metadata fragments = %+v", node.Metadata)
	}
}

func TestWriteJSONOmitsClosingSourceForUnclosedBlock(t *testing.T) {
	block := &ast.DelimitedBlock{
		Source:        span(0, 1, 1, 8, 2, 4),
		OpeningSource: span(0, 1, 1, 4, 1, 5),
		ContentSource: span(5, 2, 1, 8, 2, 4),
		Kind:          ast.DelimitedBlockListing,
		ContentModel:  ast.ContentModelVerbatim,
		Marker:        "----",
		Content:       "abc",
	}
	result := asciidoc.Result{Document: &ast.Document{
		Source: span(0, 1, 1, 8, 2, 4),
		Blocks: []ast.Block{block},
	}}

	var output bytes.Buffer
	if err := format.WriteJSON(&output, result); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	document := decoded["document"].(map[string]any)
	node := document["blocks"].([]any)[0].(map[string]any)
	if _, exists := node["closingSource"]; exists {
		t.Fatalf("unclosed block contains closingSource: %s", output.String())
	}
	if closed, ok := node["closed"].(bool); !ok || closed {
		t.Fatalf("closed = %#v, want false", node["closed"])
	}
}

func TestWriteJSONSerializesExtendedParserNodes(t *testing.T) {
	result := asciidoc.Parse([]byte(":toc:\n\n* one\n** nested\n\nTerm:: description\n\nTIP: take care\n\nimage::diagram.svg[Diagram]\n\nVisit https://example.com[Home] and <<term,Term>>."))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	var output bytes.Buffer
	if err := format.WriteJSON(&output, result); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	document := decoded["document"].(map[string]any)
	blocks := document["blocks"].([]any)
	if len(blocks) != 6 {
		t.Fatalf("blocks = %#v", blocks)
	}
	attribute := blocks[0].(map[string]any)
	if attribute["kind"] != "attribute_entry" || attribute["operation"] != "set" || attribute["name"] != "toc" || attribute["header"] != true {
		t.Fatalf("attribute = %#v", attribute)
	}
	list := blocks[1].(map[string]any)
	if list["kind"] != "list" || list["listKind"] != "unordered" || len(list["items"].([]any)) != 1 {
		t.Fatalf("list = %#v", list)
	}
	item := list["items"].([]any)[0].(map[string]any)
	if item["kind"] != "list_item" || len(item["blocks"].([]any)) != 1 {
		t.Fatalf("list item = %#v", item)
	}
	description := blocks[2].(map[string]any)
	if description["kind"] != "description_list" || description["items"].([]any)[0].(map[string]any)["term"] != "Term" {
		t.Fatalf("description list = %#v", description)
	}
	admonition := blocks[3].(map[string]any)
	if admonition["kind"] != "admonition" || admonition["admonitionKind"] != "tip" {
		t.Fatalf("admonition = %#v", admonition)
	}
	macro := blocks[4].(map[string]any)
	if macro["kind"] != "block_macro" || macro["name"] != "image" || macro["target"] != "diagram.svg" {
		t.Fatalf("macro = %#v", macro)
	}
	paragraph := blocks[5].(map[string]any)
	inlines := paragraph["inlines"].([]any)
	if inlines[1].(map[string]any)["kind"] != "link" || inlines[3].(map[string]any)["kind"] != "cross_reference" {
		t.Fatalf("paragraph inlines = %#v", inlines)
	}
}

func TestWriteJSONRejectsNilDocumentAndPropagatesWriterError(t *testing.T) {
	if err := format.WriteJSON(&bytes.Buffer{}, asciidoc.Result{}); err == nil || !strings.Contains(err.Error(), "nil document") {
		t.Fatalf("nil document error = %v", err)
	}

	want := errors.New("broken writer")
	result := asciidoc.Result{Document: &ast.Document{Source: span(0, 1, 1, 0, 1, 1)}}
	if err := format.WriteJSON(errorWriter{err: want}, result); !errors.Is(err, want) {
		t.Fatalf("writer error = %v, want %v", err, want)
	}
}

type errorWriter struct{ err error }

func (w errorWriter) Write([]byte) (int, error) { return 0, w.err }

type wireResult struct {
	Document    wireDocument     `json:"document"`
	Diagnostics []wireDiagnostic `json:"diagnostics"`
}

type wireDocument struct {
	Kind   string             `json:"kind"`
	Source wireSpan           `json:"source"`
	Title  *wireDocumentTitle `json:"title"`
	Blocks []wireNode         `json:"blocks"`
}

type wireDocumentTitle struct {
	Kind        string   `json:"kind"`
	Source      wireSpan `json:"source"`
	TitleSource wireSpan `json:"titleSource"`
	Text        string   `json:"text"`
}

type wireNode struct {
	Kind          string             `json:"kind"`
	Source        wireSpan           `json:"source"`
	HeadingSource *wireSpan          `json:"headingSource"`
	TitleSource   *wireSpan          `json:"titleSource"`
	Level         *int               `json:"level"`
	Title         *string            `json:"title"`
	Text          *string            `json:"text"`
	Inlines       *[]wireInline      `json:"inlines"`
	Blocks        *[]wireNode        `json:"blocks"`
	OpeningSource *wireSpan          `json:"openingSource"`
	ContentSource *wireSpan          `json:"contentSource"`
	ClosingSource *wireSpan          `json:"closingSource"`
	BlockKind     *string            `json:"blockKind"`
	ContentModel  *string            `json:"contentModel"`
	Marker        *string            `json:"marker"`
	Metadata      *wireBlockMetadata `json:"metadata"`
	Content       *string            `json:"content"`
	Closed        *bool              `json:"closed"`
}

type wireInline struct {
	Kind          string        `json:"kind"`
	Source        wireSpan      `json:"source"`
	ContentSource *wireSpan     `json:"contentSource"`
	Value         *string       `json:"value"`
	Children      *[]wireInline `json:"children"`
}

type wireBlockMetadata struct {
	Source         *wireSpan           `json:"source"`
	Title          *wireBlockTitle     `json:"title"`
	Anchor         *wireAnchor         `json:"anchor"`
	AttributeLists []wireAttributeList `json:"attributeLists"`
}

type wireBlockTitle struct {
	Source      wireSpan `json:"source"`
	TitleSource wireSpan `json:"titleSource"`
	Text        string   `json:"text"`
}

type wireAnchor struct {
	Source        wireSpan  `json:"source"`
	IDSource      wireSpan  `json:"idSource"`
	RefTextSource *wireSpan `json:"refTextSource"`
	ID            string    `json:"id"`
	RefText       string    `json:"refText"`
}

type wireAttributeList struct {
	Source  wireSpan        `json:"source"`
	Entries []wireAttribute `json:"entries"`
}

type wireAttribute struct {
	Source wireSpan `json:"source"`
	Value  string   `json:"value"`
}

type wireDiagnostic struct {
	Severity string   `json:"severity"`
	Message  string   `json:"message"`
	Source   wireSpan `json:"source"`
}

type wireSpan struct {
	Start wirePosition `json:"start"`
	End   wirePosition `json:"end"`
}

type wirePosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

func span(startOffset, startLine, startColumn, endOffset, endLine, endColumn int) ast.Span {
	return ast.Span{
		Start: ast.Position{Offset: startOffset, Line: startLine, Column: startColumn},
		End:   ast.Position{Offset: endOffset, Line: endLine, Column: endColumn},
	}
}
