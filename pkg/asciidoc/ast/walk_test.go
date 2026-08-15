package ast

import (
	"fmt"
	"reflect"
	"testing"
)

func TestWalkVisitsEveryNodeInDepthFirstPreOrder(t *testing.T) {
	document := &Document{
		Title: &DocumentTitle{Text: "Document"},
		Blocks: []Block{
			&Section{Title: "Section", Blocks: []Block{
				&Paragraph{Inlines: []Inline{
					&Text{Value: "before"},
					&Strong{Children: []Inline{&Text{Value: "strong"}}},
				}},
				&List{Items: []*ListItem{{
					Inlines: []Inline{&Text{Value: "item"}},
					Blocks:  []Block{&Admonition{Inlines: []Inline{&Text{Value: "note"}}}},
				}}},
				&DescriptionList{Items: []*DescriptionListItem{{
					TermInlines:        []Inline{&Text{Value: "term"}},
					DescriptionInlines: []Inline{&Text{Value: "description"}},
				}}},
				&Table{Rows: []*TableRow{{Cells: []*TableCell{{
					Inlines: []Inline{&Link{Children: []Inline{&Text{Value: "link"}}}},
				}}}}},
			}},
		},
	}

	var visited []string
	Walk(document, func(node Node) bool {
		visited = append(visited, fmt.Sprintf("%T", node))
		return true
	})
	want := []string{
		"*ast.Document", "*ast.DocumentTitle", "*ast.Section",
		"*ast.Paragraph", "*ast.Text", "*ast.Strong", "*ast.Text",
		"*ast.List", "*ast.ListItem", "*ast.Text", "*ast.Admonition", "*ast.Text",
		"*ast.DescriptionList", "*ast.DescriptionListItem", "*ast.Text", "*ast.Text",
		"*ast.Table", "*ast.TableRow", "*ast.TableCell", "*ast.Link", "*ast.Text",
	}
	if !reflect.DeepEqual(visited, want) {
		t.Fatalf("visited = %#v, want %#v", visited, want)
	}
}

func TestWalkCanPruneChildren(t *testing.T) {
	document := &Document{Blocks: []Block{
		&Section{Title: "skip", Blocks: []Block{&Paragraph{}}},
		&Section{Title: "visit", Blocks: []Block{&ThematicBreak{}}},
	}}

	var visited []string
	Walk(document, func(node Node) bool {
		visited = append(visited, fmt.Sprintf("%T", node))
		section, ok := node.(*Section)
		return !ok || section.Title != "skip"
	})
	want := []string{"*ast.Document", "*ast.Section", "*ast.Section", "*ast.ThematicBreak"}
	if !reflect.DeepEqual(visited, want) {
		t.Fatalf("visited = %#v, want %#v", visited, want)
	}
}

func TestWalkAcceptsNilRootVisitorAndChildren(t *testing.T) {
	Walk(nil, func(Node) bool { t.Fatal("visitor called for nil root"); return true })
	Walk(&Document{}, nil)

	var nilSection *Section
	document := &Document{Blocks: []Block{nilSection}}
	count := 0
	Walk(document, func(Node) bool { count++; return true })
	if count != 1 {
		t.Fatalf("visited %d nodes, want only the document", count)
	}
}
