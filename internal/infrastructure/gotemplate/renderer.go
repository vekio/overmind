// Package gotemplate implements template rendering with Go's text/template.
package gotemplate

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"text/template"

	"git.casta.me/alberto/overmind/internal/ports"
)

var _ ports.Renderer = (*Renderer)(nil)

//go:embed templates/*.tmpl
var templateFS embed.FS

// Renderer renders the application's embedded templates.
type Renderer struct {
	templates *template.Template
}

// New parses the embedded templates and creates a renderer.
func New() (*Renderer, error) {
	templates, err := template.New("documents").
		Funcs(funcMap()).
		Option("missingkey=error").
		ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse document templates: %w", err)
	}

	return &Renderer{templates: templates}, nil
}

// Render renders name using data.
func (renderer *Renderer) Render(ctx context.Context, name string, data any) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var output bytes.Buffer
	if err := renderer.templates.ExecuteTemplate(&output, name, data); err != nil {
		return nil, fmt.Errorf("render template %q: %w", name, err)
	}

	return output.Bytes(), nil
}
