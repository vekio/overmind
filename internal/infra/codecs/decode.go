package codecs

import (
	"bytes"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
	"github.com/vekio/overmind/pkg/asciidoc"
	"github.com/vekio/overmind/pkg/asciidoc/semantic"
)

type noteHeader struct {
	header   semantic.Header
	id       uuid.UUID
	kind     string
	tags     shared.Tags
	metadata shared.EntityMetadata
	body     string
}

// splitDocument keeps the original body bytes, including leading blank lines
// and trailing whitespace, while limiting metadata parsing to the header.
func splitDocument(source []byte) (header []byte, body []byte) {
	started := false
	for offset := 0; offset < len(source); {
		end := bytes.IndexByte(source[offset:], '\n')
		if end < 0 {
			break
		}
		end += offset
		line := source[offset:end]
		if len(bytes.TrimSpace(line)) == 0 && started {
			return source[:offset], source[end+1:]
		}
		if len(bytes.TrimSpace(line)) != 0 {
			started = true
		}
		offset = end + 1
	}
	return source, nil
}

func decodeHeader(source []byte, expected string) (noteHeader, error) {
	headerSource, body := splitDocument(source)
	result := asciidoc.Process(headerSource)
	if result.HasErrors() {
		return noteHeader{}, fmt.Errorf("invalid AsciiDoc note header: %v", result.Diagnostics)
	}
	header := result.Analysis.Header
	kind, err := required(header.Attributes, "overmind-type")
	if err != nil {
		return noteHeader{}, err
	}
	if expected != "" && kind != expected {
		return noteHeader{}, fmt.Errorf("expected %s note, found %q", expected, kind)
	}
	text, err := required(header.Attributes, "overmind-id")
	if err != nil {
		return noteHeader{}, err
	}
	id, err := uuid.Parse(text)
	if err != nil {
		return noteHeader{}, fmt.Errorf("invalid note ID: %w", err)
	}
	if id == uuid.Nil() {
		return noteHeader{}, fmt.Errorf("note ID must not be nil")
	}
	tags, err := parseTags(header.Attributes)
	if err != nil {
		return noteHeader{}, err
	}
	metadata, err := parseMetadata(header.Attributes)
	if err != nil {
		return noteHeader{}, err
	}
	return noteHeader{header: header, id: id, kind: kind, tags: tags, metadata: metadata, body: string(body)}, nil
}

// Identify reads the identity and kind used to select a document's codec.
// Specific domain properties are validated by that codec's Decode method.
func Identify(source []byte) (uuid.UUID, ports.NoteKind, error) {
	note, err := decodeHeader(source, "")
	if err != nil {
		return uuid.Nil(), "", err
	}
	return note.id, ports.NoteKind(note.kind), nil
}
