package ast

// Node is any element in the parsed document tree.
type Node interface {
	SourceSpan() Span
}

// Block is a node that may occur in a document or section body.
type Block interface {
	Node
	blockNode()
}
