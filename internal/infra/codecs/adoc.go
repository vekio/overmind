// Package codecs maps domain entities to managed AsciiDoc documents.
// Typed codecs validate headers and domain values, keeping body bytes verbatim.
// RawNoteCodec additionally checks complete document syntax before raw saves.
package codecs

import (
	"bytes"
	"embed"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
	"github.com/vekio/overmind/pkg/asciidoc/semantic"
)

//go:embed templates/*.adoc.tmpl
var sources embed.FS
var templates = template.Must(template.New("notes").Funcs(template.FuncMap{
	"tags":         func(tags shared.Tags) string { return strings.Join(tags.Strings(), ", ") },
	"groups":       func(groups persons.Groups) string { return strings.Join(groups.Strings(), ", ") },
	"timestamp":    func(at time.Time) string { return at.UTC().Format(time.RFC3339Nano) },
	"number":       func(value float64) string { return strconv.FormatFloat(value, 'g', -1, 64) },
	"habitKind":    func() string { return ports.NoteKindHabit.String() },
	"personKind":   func() string { return ports.NoteKindPerson.String() },
	"bookmarkKind": func() string { return ports.NoteKindBookmark.String() },
	"inboxKind":    func() string { return ports.NoteKindInbox.String() },
	"pageKind":     func() string { return ports.NoteKindPage.String() },
	"journalKind":  func() string { return ports.NoteKindJournal.String() },
	"journalTitle": journalTitle,
}).ParseFS(sources, "templates/*.adoc.tmpl"))

func render(name string, value any) ([]byte, error) {
	var output bytes.Buffer
	if err := templates.ExecuteTemplate(&output, name, value); err != nil {
		return nil, fmt.Errorf("render %s: %w", name, err)
	}
	return output.Bytes(), nil
}
func required(attributes semantic.AttributeSet, name string) (string, error) {
	value, ok := attributes.Lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}
func journalTitle(value string) (string, error) {
	date, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return "", fmt.Errorf("parse journal date: %w", err)
	}
	months := [...]string{"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio", "Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"}
	return fmt.Sprintf("%s %d, %d", months[date.Month()-1], date.Day(), date.Year()), nil
}

func parseTags(attributes semantic.AttributeSet) (shared.Tags, error) {
	var values []shared.Tag
	if text, ok := attributes.Lookup("overmind-tags"); ok && strings.TrimSpace(text) != "" {
		for _, part := range strings.Split(text, ",") {
			tag, err := shared.NewTag(part)
			if err != nil {
				return shared.Tags{}, err
			}
			values = append(values, tag)
		}
	}
	return shared.NewTags(values...)
}
func parseMetadata(attributes semantic.AttributeSet) (shared.EntityMetadata, error) {
	createdText, err := required(attributes, "overmind-created-at")
	if err != nil {
		return shared.EntityMetadata{}, err
	}
	created, err := time.Parse(time.RFC3339Nano, createdText)
	if err != nil {
		return shared.EntityMetadata{}, fmt.Errorf("invalid creation time: %w", err)
	}
	updatedText, err := required(attributes, "overmind-updated-at")
	if err != nil {
		return shared.EntityMetadata{}, err
	}
	updated, err := time.Parse(time.RFC3339Nano, updatedText)
	if err != nil {
		return shared.EntityMetadata{}, fmt.Errorf("invalid update time: %w", err)
	}
	return shared.NewEntityMetadata(created, updated)
}
