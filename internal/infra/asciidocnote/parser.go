// Package asciidocnote reads metadata from Overmind AsciiDoc notes.
package asciidocnote

import (
	"fmt"
	"strings"
	"time"
	"uuid"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
	"git.casta.me/alberto/overmind/pkg/asciidoc"
	"git.casta.me/alberto/overmind/pkg/asciidoc/diagnostic"
	"git.casta.me/alberto/overmind/pkg/asciidoc/semantic"
)

var _ ports.NoteParser = Parser{}

// Parser extracts indexable metadata with the public AsciiDoc pipeline.
type Parser struct{}

// Parse extracts and validates indexable metadata from an AsciiDoc note.
func (Parser) Parse(source []byte) (ports.IndexRecord, error) {
	result := asciidoc.Process(source)
	if result.HasErrors() {
		for _, item := range result.Diagnostics {
			if item.Severity == diagnostic.SeverityError {
				return ports.IndexRecord{}, fmt.Errorf("AsciiDoc %s", item)
			}
		}
	}
	attributes := result.Analysis.Header.Attributes

	idText, err := required(attributes, "overmind-id")
	if err != nil {
		return ports.IndexRecord{}, err
	}
	id, err := uuid.Parse(idText)
	if err != nil || id == uuid.Nil() {
		return ports.IndexRecord{}, fmt.Errorf("invalid overmind-id %q", idText)
	}
	kindText, err := required(attributes, "overmind-type")
	if err != nil {
		return ports.IndexRecord{}, err
	}
	kind, err := domain.ParseNoteKind(kindText)
	if err != nil {
		return ports.IndexRecord{}, err
	}
	createdAt, err := timestamp(attributes, "overmind-created-at")
	if err != nil {
		return ports.IndexRecord{}, err
	}
	updatedAt, err := timestamp(attributes, "overmind-updated-at")
	if err != nil {
		return ports.IndexRecord{}, err
	}
	if updatedAt.Before(createdAt) {
		return ports.IndexRecord{}, fmt.Errorf("overmind-updated-at precedes overmind-created-at")
	}
	tags, err := parseTags(attributes)
	if err != nil {
		return ports.IndexRecord{}, err
	}
	record := ports.IndexRecord{ID: id, Kind: kind, CreatedAt: createdAt, UpdatedAt: updatedAt, Tags: tags}

	switch kind {
	case domain.NoteKindPerson:
		if result.Analysis.Header.Title == nil {
			return ports.IndexRecord{}, fmt.Errorf("person name is required")
		}
		name, err := domain.NewTitle(result.Analysis.Header.Title.Text)
		if err != nil {
			return ports.IndexRecord{}, err
		}
		record.Attributes = append(record.Attributes, ports.IndexAttribute{Name: "name", Value: name.String()})
		if value, ok := attributes.Lookup("overmind-groups"); ok && strings.TrimSpace(value) != "" {
			var values []domain.Group
			for _, part := range strings.Split(value, ",") {
				group, err := domain.NewGroup(part)
				if err != nil {
					return ports.IndexRecord{}, fmt.Errorf("invalid overmind-groups: %w", err)
				}
				values = append(values, group)
			}
			groups, err := domain.NewGroups(values...)
			if err != nil {
				return ports.IndexRecord{}, fmt.Errorf("invalid overmind-groups: %w", err)
			}
			record.Attributes = append(record.Attributes, ports.IndexAttribute{Name: "groups", Value: strings.Join(groups.Strings(), ", ")})
		}
	case domain.NoteKindPage:
		if result.Analysis.Header.Title == nil {
			return ports.IndexRecord{}, fmt.Errorf("page title is required")
		}
		title, err := domain.NewTitle(result.Analysis.Header.Title.Text)
		if err != nil {
			return ports.IndexRecord{}, err
		}
		record.Attributes = append(record.Attributes, ports.IndexAttribute{Name: "title", Value: title.String()})
		if value, ok := attributes.Lookup("overmind-area"); ok && strings.TrimSpace(value) != "" {
			area, err := domain.NewArea(value)
			if err != nil {
				return ports.IndexRecord{}, err
			}
			record.Attributes = append(record.Attributes, ports.IndexAttribute{Name: "area", Value: area.String()})
		}
	case domain.NoteKindJournal:
		value, err := required(attributes, "overmind-date")
		if err != nil {
			return ports.IndexRecord{}, err
		}
		date, err := domain.NewDate(value)
		if err != nil {
			return ports.IndexRecord{}, err
		}
		record.Attributes = append(record.Attributes, ports.IndexAttribute{Name: "date", Value: date.String()})
	case domain.NoteKindBookmark:
		value, err := required(attributes, "overmind-url")
		if err != nil {
			return ports.IndexRecord{}, err
		}
		bookmarkURL, err := domain.NewURL(value)
		if err != nil {
			return ports.IndexRecord{}, err
		}
		record.Attributes = append(record.Attributes, ports.IndexAttribute{Name: "url", Value: bookmarkURL.String()})
	case domain.NoteKindInbox:
	}
	return record, nil
}

func required(attributes semantic.AttributeSet, name string) (string, error) {
	value, ok := attributes.Lookup(name)
	value = strings.TrimSpace(value)
	if !ok || value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func timestamp(attributes semantic.AttributeSet, name string) (time.Time, error) {
	value, err := required(attributes, name)
	if err != nil {
		return time.Time{}, err
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil || parsed.IsZero() {
		return time.Time{}, fmt.Errorf("invalid %s %q: expected RFC3339", name, value)
	}
	return parsed, nil
}

func parseTags(attributes semantic.AttributeSet) ([]string, error) {
	value, ok := attributes.Lookup("overmind-tags")
	if !ok || strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	tags := make([]domain.Tag, 0, len(parts))
	for _, part := range parts {
		tag, err := domain.NewTag(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("invalid overmind-tags: %w", err)
		}
		tags = append(tags, tag)
	}
	set, err := domain.NewTags(tags...)
	if err != nil {
		return nil, fmt.Errorf("invalid overmind-tags: %w", err)
	}
	return set.Strings(), nil
}
