package renderer

import (
	"fmt"
	"strings"
	"text/template"
	"time"
)

var spanishMonths = [...]string{
	"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
	"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre",
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"comma": func(values []string) string {
			return strings.Join(values, ", ")
		},
		"rfc3339": func(value time.Time) string {
			return value.UTC().Format(time.RFC3339)
		},
		"journalTitle": func(value string) (string, error) {
			date, err := time.Parse(time.DateOnly, value)
			if err != nil {
				return "", fmt.Errorf("parse journal date: %w", err)
			}
			return fmt.Sprintf("%s %d, %d", spanishMonths[date.Month()-1], date.Day(), date.Year()), nil
		},
	}
}
