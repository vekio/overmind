package renderer

import "time"

// PageTemplate identifies the AsciiDoc page template.
const PageTemplate = "page"

// Page contains the data used by PageTemplate.
type Page struct {
	ID        string
	Title     string
	Area      string
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}
