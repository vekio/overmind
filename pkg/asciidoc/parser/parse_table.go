package parser

import (
	"fmt"
	"strconv"
	"strings"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/diagnostic"
	"git.casta.me/alberto/overmind/pkg/asciidoc/lexer"
)

func (p *Parser) parseTable(metadata ast.BlockMetadata) ast.Block {
	opening := p.current
	openingSource := tokenSpan(opening)
	contentStart := tokenEndPosition(opening)
	format, separator := tableFormat(opening.Delimiter.Marker)
	table := &ast.Table{
		Source:        sourceWithMetadata(ast.Span{Start: openingSource.Start}, metadata),
		OpeningSource: openingSource,
		ContentSource: ast.Span{Start: contentStart, End: contentStart},
		Marker:        opening.Delimiter.Marker,
		Format:        format,
		Separator:     separator,
		Metadata:      metadata,
	}
	configuration := tableConfiguration(metadata)
	table.Columns = configuration.columns
	table.Header = configuration.header
	if configuration.formatSpecified {
		table.Format = configuration.format
	}
	if configuration.separator != "" {
		table.Separator = configuration.separator
	}

	p.advance()
	var content strings.Builder
	for p.current.Kind != lexer.LineEOF {
		if p.current.Kind == lexer.LineDelimiter && p.current.Delimiter.Marker == table.Marker {
			table.ClosingSource = tokenSpan(p.current)
			table.ContentSource.End = table.ClosingSource.Start
			table.Source.End = table.ClosingSource.End
			table.Content = content.String()
			table.Closed = true
			p.advance()
			p.parseTableContent(table)
			return table
		}
		content.WriteString(p.current.SourceText())
		p.advance()
	}

	eof := tokenSpan(p.current).Start
	table.ContentSource.End = eof
	table.Source.End = eof
	table.Content = content.String()
	p.appendDiagnostic(diagnostic.SeverityError, fmt.Sprintf("unclosed table; expected %q", table.Marker), openingSource)
	p.parseTableContent(table)
	return table
}

func (p *Parser) parseTableContent(table *ast.Table) {
	if table.Format != ast.TableFormatPSV || len(table.Separator) != 1 {
		p.appendDiagnostic(diagnostic.SeverityWarning, fmt.Sprintf("parser does not yet support %s table data", table.Format), table.OpeningSource)
		return
	}
	separator := table.Separator[0]
	cells := parsePSVCells(table.Content, table.ContentSource.Start, separator)
	columnCount := len(table.Columns)
	if columnCount == 0 {
		columnCount = inferTableColumnCount(table.Content, separator)
	}
	if columnCount == 0 && len(cells) > 0 {
		columnCount = 1
	}
	if len(table.Columns) == 0 {
		table.Columns = make([]ast.TableColumn, columnCount)
	}
	for start := 0; start < len(cells); start += columnCount {
		end := min(start+columnCount, len(cells))
		row := &ast.TableRow{Cells: cells[start:end]}
		row.Source = ast.Span{Start: row.Cells[0].Source.Start, End: row.Cells[len(row.Cells)-1].Source.End}
		table.Rows = append(table.Rows, row)
	}
	if len(table.Rows) > 0 {
		if !table.Header {
			table.Header = inferHeaderRow(table.Content, columnCount, separator)
		}
		table.Rows[0].Header = table.Header
	}
	if columnCount > 0 && len(cells)%columnCount != 0 {
		p.appendDiagnostic(diagnostic.SeverityWarning, fmt.Sprintf("table row has %d cells; expected %d", len(cells)%columnCount, columnCount), table.Rows[len(table.Rows)-1].Source)
	}
}

func parsePSVCells(source string, origin ast.Position, separator byte) []*ast.TableCell {
	positions := inlinePositions(source, origin)
	separators := psvSeparators(source, separator)
	result := make([]*ast.TableCell, 0, len(separators))
	for index, separator := range separators {
		end := len(source)
		if index+1 < len(separators) {
			end = separators[index+1]
		}
		contentStart, contentEnd := trimTableCell(source, separator+1, end)
		text := normalizeTableText(source[contentStart:contentEnd])
		result = append(result, &ast.TableCell{
			Source:        ast.Span{Start: positions[separator], End: positions[contentEnd]},
			MarkerSource:  ast.Span{Start: positions[separator], End: positions[separator+1]},
			ContentSource: ast.Span{Start: positions[contentStart], End: positions[contentEnd]},
			Text:          text,
			Inlines:       parseInlines(source[contentStart:contentEnd], positions[contentStart]),
		})
	}
	return result
}

func psvSeparators(source string, separator byte) []int {
	var result []int
	lineStart := true
	for index := 0; index < len(source); index++ {
		if source[index] == '\\' && index+1 < len(source) && (source[index+1] == separator || source[index+1] == '\\') {
			index++
			lineStart = false
			continue
		}
		if source[index] == separator && (lineStart || index > 0 && (source[index-1] == ' ' || source[index-1] == '\t')) {
			result = append(result, index)
		}
		if source[index] == '\n' || source[index] == '\r' {
			lineStart = true
			if source[index] == '\r' && index+1 < len(source) && source[index+1] == '\n' {
				index++
			}
			continue
		}
		lineStart = false
	}
	return result
}

func trimTableCell(source string, start, end int) (int, int) {
	for start < end && (source[start] == ' ' || source[start] == '\t') {
		start++
	}
	for end > start && strings.ContainsRune(" \t\r\n", rune(source[end-1])) {
		end--
	}
	return start, end
}

func normalizeTableText(source string) string {
	return strings.ReplaceAll(strings.ReplaceAll(source, "\r\n", "\n"), "\r", "\n")
}

func inferTableColumnCount(source string, separator byte) int {
	end := strings.IndexAny(source, "\r\n")
	if end < 0 {
		end = len(source)
	}
	return len(psvSeparators(source[:end], separator))
}

func inferHeaderRow(source string, columnCount int, separator byte) bool {
	if columnCount == 0 {
		return false
	}
	lineEnd := strings.IndexAny(source, "\r\n")
	if lineEnd < 0 {
		return false
	}
	endingWidth := 1
	if source[lineEnd] == '\r' && lineEnd+1 < len(source) && source[lineEnd+1] == '\n' {
		endingWidth = 2
	}
	firstLine := source[:lineEnd]
	remaining := source[lineEnd+endingWidth:]
	return len(psvSeparators(firstLine, separator)) == columnCount && (strings.HasPrefix(remaining, "\n") || strings.HasPrefix(remaining, "\r\n") || strings.HasPrefix(remaining, "\r"))
}

type tableConfig struct {
	columns         []ast.TableColumn
	header          bool
	format          ast.TableFormat
	separator       string
	formatSpecified bool
}

func tableConfiguration(metadata ast.BlockMetadata) tableConfig {
	var result tableConfig
	for _, list := range metadata.AttributeLists {
		for _, entry := range list.Entries {
			value := entry.Value
			if hasTableHeaderOption(value) {
				result.header = true
			}
			if strings.HasPrefix(value, "cols=") {
				result.columns = parseTableColumns(attributeValue(value))
			}
			if strings.HasPrefix(value, "format=") {
				result.formatSpecified = true
				result.format = namedTableFormat(attributeValue(value))
			}
			if strings.HasPrefix(value, "separator=") {
				result.separator = attributeValue(value)
			}
		}
	}
	return result
}

func hasTableHeaderOption(value string) bool {
	if value == "%header" || value == "header" {
		return true
	}
	if !strings.HasPrefix(value, "options=") {
		return false
	}
	for _, option := range strings.Split(attributeValue(value), ",") {
		if strings.TrimSpace(option) == "header" {
			return true
		}
	}
	return false
}

func attributeValue(entry string) string {
	_, value, _ := strings.Cut(entry, "=")
	return strings.Trim(value, "\"'")
}

func namedTableFormat(value string) ast.TableFormat {
	switch value {
	case "psv":
		return ast.TableFormatPSV
	case "csv":
		return ast.TableFormatCSV
	case "dsv":
		return ast.TableFormatDSV
	case "tsv":
		return ast.TableFormatTSV
	default:
		return ast.TableFormatUnknown
	}
}

func parseTableColumns(value string) []ast.TableColumn {
	if value == "" {
		return nil
	}
	if strings.HasSuffix(value, "*") {
		if count, err := strconv.Atoi(strings.TrimSuffix(value, "*")); err == nil && count > 0 {
			return make([]ast.TableColumn, count)
		}
	}
	parts := strings.Split(value, ",")
	result := make([]ast.TableColumn, 0, len(parts))
	for _, part := range parts {
		result = append(result, ast.TableColumn{Spec: strings.TrimSpace(part)})
	}
	return result
}

func tableFormat(marker string) (ast.TableFormat, string) {
	switch marker[0] {
	case '|', '!':
		return ast.TableFormatPSV, marker[:1]
	case ',':
		return ast.TableFormatCSV, ","
	case ':':
		return ast.TableFormatDSV, ":"
	default:
		return ast.TableFormatUnknown, ""
	}
}
