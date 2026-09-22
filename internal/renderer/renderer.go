// Package renderer renders embedded document templates.
package renderer

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// Renderer renders embedded templates.
type Renderer struct {
	templates *template.Template
}

// New creates a Renderer.
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

// Render executes a template with data.
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
