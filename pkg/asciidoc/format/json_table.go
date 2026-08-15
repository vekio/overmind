package format

import "git.casta.me/alberto/overmind/pkg/asciidoc/ast"

func makeJSONTable(table *ast.Table) (jsonTable, error) {
	result := jsonTable{
		Kind:          "table",
		Source:        makeJSONSpan(table.Source),
		OpeningSource: makeJSONSpan(table.OpeningSource),
		ContentSource: makeJSONSpan(table.ContentSource),
		Marker:        table.Marker,
		Format:        table.Format.String(),
		Separator:     table.Separator,
		Metadata:      makeJSONBlockMetadata(table.Metadata),
		Columns:       make([]jsonTableColumn, 0, len(table.Columns)),
		Rows:          make([]jsonTableRow, 0, len(table.Rows)),
		Content:       table.Content,
		Header:        table.Header,
		Closed:        table.Closed,
	}
	if table.Closed {
		closing := makeJSONSpan(table.ClosingSource)
		result.ClosingSource = &closing
	}
	for _, column := range table.Columns {
		result.Columns = append(result.Columns, jsonTableColumn{Spec: column.Spec})
	}
	for _, row := range table.Rows {
		encodedRow := jsonTableRow{
			Kind:   "table_row",
			Source: makeJSONSpan(row.Source),
			Header: row.Header,
			Cells:  make([]jsonTableCell, 0, len(row.Cells)),
		}
		for _, cell := range row.Cells {
			inlines, err := makeJSONInlines(cell.Inlines)
			if err != nil {
				return jsonTable{}, err
			}
			encodedRow.Cells = append(encodedRow.Cells, jsonTableCell{
				Kind:          "table_cell",
				Source:        makeJSONSpan(cell.Source),
				MarkerSource:  makeJSONSpan(cell.MarkerSource),
				ContentSource: makeJSONSpan(cell.ContentSource),
				Text:          cell.Text,
				Inlines:       inlines,
			})
		}
		result.Rows = append(result.Rows, encodedRow)
	}
	return result, nil
}
