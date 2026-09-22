package renderer

import (
	"strings"
	"text/template"
	"time"
)

func funcMap() template.FuncMap {
	return template.FuncMap{
		"comma": func(values []string) string {
			return strings.Join(values, ", ")
		},
		"rfc3339": func(value time.Time) string {
			return value.UTC().Format(time.RFC3339)
		},
	}
}
