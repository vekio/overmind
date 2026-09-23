package renderer

import "time"

// PageTemplate identifies the AsciiDoc page template.
const PageTemplate = "page"

// JournalTemplate identifies the AsciiDoc journal template.
const JournalTemplate = "journal"

// InboxTemplate identifies the AsciiDoc inbox template.
const InboxTemplate = "inbox"

// BookmarkTemplate identifies the AsciiDoc bookmark template.
const BookmarkTemplate = "bookmark"

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

// Inbox contains the data used by InboxTemplate.
type Inbox struct {
	ID        string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Bookmark contains the data used by BookmarkTemplate.
type Bookmark struct {
	ID        string
	Title     string
	URL       string
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}
