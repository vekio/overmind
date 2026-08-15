// Package format serializes parser results into interchange formats.
package format

import (
	"encoding/json"
	"io"

	"git.casta.me/alberto/overmind/pkg/asciidoc"
)

// WriteJSON writes a syntax parser result as indented JSON followed by a
// newline. It does not serialize semantic Analysis or ProcessResult values.
func WriteJSON(writer io.Writer, result asciidoc.Result) error {
	output, err := makeJSONResult(result)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}
