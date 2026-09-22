package renderer

import "time"

// PageTemplate identifies the AsciiDoc page template.
const PageTemplate = "page"

// JournalTemplate identifies the AsciiDoc journal template.
const JournalTemplate = "journal"

// Page contains the data used by PageTemplate.
type Page struct {
	ID        string
	Title     string
	Area      string
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Journal contains the data used by JournalTemplate.
type Journal struct {
	ID        string
	Date      string
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}
