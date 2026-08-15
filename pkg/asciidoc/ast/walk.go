package ast

import "reflect"

// Visitor is called for each node visited by Walk. Returning false prevents
// Walk from descending into that node's children; it does not stop traversal
// of the node's siblings.
type Visitor func(Node) bool

// Walk traverses root and its descendants in depth-first pre-order. Block
// metadata is not visited because it is not represented by Node values.
// Walk does nothing when root or visitor is nil.
func Walk(root Node, visitor Visitor) {
	if isNilNode(root) || visitor == nil || !visitor(root) {
		return
	}

	switch node := root.(type) {
	case *Document:
		walkNode(node.Title, visitor)
		walkBlocks(node.Blocks, visitor)
	case *Section:
		walkBlocks(node.Blocks, visitor)
	case *Paragraph:
		walkInlines(node.Inlines, visitor)
	case *Strong:
		walkInlines(node.Children, visitor)
	case *Emphasis:
		walkInlines(node.Children, visitor)
	case *Monospace:
		walkInlines(node.Children, visitor)
	case *Link:
		walkInlines(node.Children, visitor)
	case *CrossReference:
		walkInlines(node.Children, visitor)
	case *List:
		for _, item := range node.Items {
			walkNode(item, visitor)
		}
	case *ListItem:
		walkInlines(node.Inlines, visitor)
		walkBlocks(node.Blocks, visitor)
	case *DescriptionList:
		for _, item := range node.Items {
			walkNode(item, visitor)
		}
	case *DescriptionListItem:
		walkInlines(node.TermInlines, visitor)
		walkInlines(node.DescriptionInlines, visitor)
		walkBlocks(node.Blocks, visitor)
	case *Admonition:
		walkInlines(node.Inlines, visitor)
	case *Table:
		for _, row := range node.Rows {
			walkNode(row, visitor)
		}
	case *TableRow:
		for _, cell := range node.Cells {
			walkNode(cell, visitor)
		}
	case *TableCell:
		walkInlines(node.Inlines, visitor)
	}
}

func walkBlocks(blocks []Block, visitor Visitor) {
	for _, block := range blocks {
		walkNode(block, visitor)
	}
}

func walkInlines(inlines []Inline, visitor Visitor) {
	for _, inline := range inlines {
		walkNode(inline, visitor)
	}
}

func walkNode(node Node, visitor Visitor) {
	if !isNilNode(node) {
		Walk(node, visitor)
	}
}

func isNilNode(node Node) bool {
	if node == nil {
		return true
	}
	value := reflect.ValueOf(node)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	}
	return false
}
