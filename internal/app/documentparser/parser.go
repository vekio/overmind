// Package documentparser interprets metadata from Overmind AsciiDoc sources.
package documentparser

import (
	"fmt"
	"strings"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/pkg/asciidoc"
)

const (
	headerPrefix    = "overmind-"
	headerID        = headerPrefix + "id"
	headerKind      = headerPrefix + "type"
	headerArea      = headerPrefix + "area"
	headerTags      = headerPrefix + "tags"
	headerCreatedAt = headerPrefix + "created-at"
	headerUpdatedAt = headerPrefix + "updated-at"
)

// ParsedDocument contains validated metadata read from an AsciiDoc source.
// It is independent from storage and from the document index representation.
type ParsedDocument struct {
	ID         domain.DocumentID
	Kind       domain.DocumentKind
	Title      domain.Title
	Area       domain.Area
	Tags       domain.Tags
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Attributes map[string]string
}

// Parse interprets one AsciiDoc source. Sources without an Overmind ID are
// reported as unmanaged and may be ignored by callers such as index rebuilds.
func Parse(source []byte) (ParsedDocument, bool, error) {
	processed := asciidoc.Process(source)
	headers := overmindHeaders(processed)
	idValue, managed := headers[headerID]
	if !managed {
		return ParsedDocument{}, false, nil
	}
	if processed.HasErrors() {
		return ParsedDocument{}, true, fmt.Errorf("invalid AsciiDoc: %s", processed.Diagnostics[0])
	}

	documentID, err := domain.NewDocumentID(idValue)
	if err != nil {
		return ParsedDocument{}, true, err
	}
	kind, err := domain.NewDocumentKind(headers[headerKind])
	if err != nil {
		return ParsedDocument{}, true, fmt.Errorf("invalid %s: %w", headerKind, err)
	}
	titleValue := ""
	if processed.Analysis.Header.Title != nil {
		titleValue = processed.Analysis.Header.Title.Text
	}
	title, err := domain.NewTitle(titleValue)
	if err != nil {
		return ParsedDocument{}, true, fmt.Errorf("invalid document title: %w", err)
	}
	tags, err := tagsFromHeader(headers[headerTags])
	if err != nil {
		return ParsedDocument{}, true, fmt.Errorf("invalid %s: %w", headerTags, err)
	}
	area, err := areaFromHeader(headers[headerArea])
	if err != nil {
		return ParsedDocument{}, true, fmt.Errorf("invalid %s: %w", headerArea, err)
	}
	createdAt, err := time.Parse(time.RFC3339, headers[headerCreatedAt])
	if err != nil {
		return ParsedDocument{}, true, fmt.Errorf("invalid %s: %w", headerCreatedAt, err)
	}
	updatedAt, err := time.Parse(time.RFC3339, headers[headerUpdatedAt])
	if err != nil {
		return ParsedDocument{}, true, fmt.Errorf("invalid %s: %w", headerUpdatedAt, err)
	}

	return ParsedDocument{
		ID:         documentID,
		Kind:       kind,
		Title:      title,
		Area:       area,
		Tags:       tags,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Attributes: documentAttributes(headers),
	}, true, nil
}

func overmindHeaders(processed asciidoc.ProcessResult) map[string]string {
	headers := make(map[string]string)
	for _, attribute := range processed.Analysis.Header.Attributes.All() {
		if strings.HasPrefix(attribute.Name, headerPrefix) {
			headers[attribute.Name] = attribute.Value
		}
	}
	return headers
}

func documentAttributes(headers map[string]string) map[string]string {
	attributes := make(map[string]string)
	for name, value := range headers {
		switch name {
		case headerID, headerKind, headerArea, headerTags, headerCreatedAt, headerUpdatedAt:
			continue
		default:
			attributes[strings.TrimPrefix(name, headerPrefix)] = value
		}
	}
	return attributes
}

func areaFromHeader(value string) (domain.Area, error) {
	if strings.TrimSpace(value) == "" {
		return domain.Area{}, nil
	}
	return domain.NewArea(value)
}

func tagsFromHeader(value string) (domain.Tags, error) {
	if strings.TrimSpace(value) == "" {
		return domain.Tags{}, nil
	}
	values := strings.Split(value, ",")
	tags := make([]domain.Tag, len(values))
	for index := range values {
		tag, err := domain.NewTag(strings.TrimSpace(values[index]))
		if err != nil {
			return domain.Tags{}, err
		}
		tags[index] = tag
	}
	return domain.NewTags(tags...)
}
