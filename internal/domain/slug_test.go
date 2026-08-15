package domain

import "testing"

func TestSlugify(t *testing.T) {
	for input, want := range map[string]string{
		"Mi primera página":       "mi-primera-pagina",
		"Über café":               "uber-cafe",
		"  spaces -- together  ":  "spaces-together",
		"Page 42":                 "page-42",
		"already-normalized":      "already-normalized",
		"symbols !@#$ separators": "symbols-separators",
		"":                        "",
		"---":                     "",
	} {
		t.Run(input, func(t *testing.T) {
			if got := Slugify(input); got != want {
				t.Fatalf("Slugify(%q) = %q, want %q", input, got, want)
			}
		})
	}
}
