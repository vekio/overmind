package ast

import "testing"

func TestPositionSpanAndNodes(t *testing.T) {
	start := Position{Offset: 4, Line: 2, Column: 3}
	end := Position{Offset: 9, Line: 2, Column: 8}
	span := Span{Start: start, End: end}
	if start.String() != "2:3@4" || span.String() != "[2:3@4, 2:8@9)" {
		t.Fatalf("position=%s span=%s", start, span)
	}

	nodes := []Node{
		&Document{Source: span},
		&DocumentTitle{Source: span},
		&Section{Source: span},
		&Paragraph{Source: span},
		&ThematicBreak{Source: span},
		&PageBreak{Source: span},
		&DelimitedBlock{Source: span},
		&Text{Source: span},
		&Strong{Source: span},
		&Emphasis{Source: span},
		&Monospace{Source: span},
		&Link{Source: span},
		&CrossReference{Source: span},
		&List{Source: span},
		&DescriptionList{Source: span},
		&Admonition{Source: span},
		&BlockMacro{Source: span},
		&AttributeEntry{Source: span},
		&Table{Source: span},
		&TableRow{Source: span},
		&TableCell{Source: span},
	}
	for _, node := range nodes {
		if node.SourceSpan() != span {
			t.Errorf("%T.SourceSpan() = %s, want %s", node, node.SourceSpan(), span)
		}
	}
}

func TestNewKindsString(t *testing.T) {
	for _, test := range []struct {
		value interface{ String() string }
		want  string
	}{
		{ListUnknown, "unknown"},
		{ListUnordered, "unordered"},
		{ListOrdered, "ordered"},
		{ListKind(255), "unknown"},
		{AdmonitionUnknown, "unknown"},
		{AdmonitionNote, "note"},
		{AdmonitionTip, "tip"},
		{AdmonitionImportant, "important"},
		{AdmonitionCaution, "caution"},
		{AdmonitionWarning, "warning"},
		{AdmonitionKind(255), "unknown"},
		{AttributeOperationUnknown, "unknown"},
		{AttributeSet, "set"},
		{AttributeUnset, "unset"},
		{AttributeOperation(255), "unknown"},
		{TableFormatUnknown, "unknown"},
		{TableFormatPSV, "psv"},
		{TableFormatCSV, "csv"},
		{TableFormatDSV, "dsv"},
		{TableFormatTSV, "tsv"},
		{TableFormat(255), "unknown"},
	} {
		if got := test.value.String(); got != test.want {
			t.Errorf("String() = %q, want %q", got, test.want)
		}
	}
}

func TestDelimitedBlockKindAndContentModelString(t *testing.T) {
	for _, test := range []struct {
		value interface{ String() string }
		want  string
	}{
		{DelimitedBlockUnknown, "unknown"},
		{DelimitedBlockOpen, "open"},
		{DelimitedBlockListing, "listing"},
		{DelimitedBlockLiteral, "literal"},
		{DelimitedBlockExample, "example"},
		{DelimitedBlockSidebar, "sidebar"},
		{DelimitedBlockQuote, "quote"},
		{DelimitedBlockPassthrough, "passthrough"},
		{DelimitedBlockComment, "comment"},
		{DelimitedBlockKind(255), "unknown"},
		{ContentModelUnknown, "unknown"},
		{ContentModelCompound, "compound"},
		{ContentModelVerbatim, "verbatim"},
		{ContentModelRaw, "raw"},
		{ContentModel(255), "unknown"},
	} {
		if got := test.value.String(); got != test.want {
			t.Errorf("String() = %q, want %q", got, test.want)
		}
	}
}
