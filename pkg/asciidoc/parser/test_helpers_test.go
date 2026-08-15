package parser_test

import (
	"bytes"
	"io"
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	parserpkg "git.casta.me/alberto/overmind/pkg/asciidoc/parser"
)

func parse(source []byte) parserpkg.Result {
	return parseReader(bytes.NewReader(source))
}

func parseReader(reader io.Reader) parserpkg.Result {
	return parserpkg.New(reader).Parse()
}

func requireBlock[T ast.Block](t *testing.T, block ast.Block) T {
	t.Helper()
	result, ok := block.(T)
	if !ok {
		t.Fatalf("block type = %T, want %T", block, result)
	}
	return result
}

func requireBlockText(t *testing.T, block ast.Block) string {
	t.Helper()
	return requireBlock[*ast.Paragraph](t, block).Text
}
