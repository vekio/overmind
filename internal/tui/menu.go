package tui

type action uint8

const (
	actionCapture action = iota
	actionPage
	actionBookmark
	actionJournal
	actionRebuild
	actionList
	actionPerson
)

func (a action) String() string {
	switch a {
	case actionPerson:
		return "Person"
	case actionPage:
		return "Page"
	case actionBookmark:
		return "Bookmark"
	case actionJournal:
		return "Journal"
	case actionRebuild:
		return "Rebuild index"
	case actionList:
		return "Notes"
	default:
		return "Capture"
	}
}

type menuItem struct {
	action      action
	title       string
	description string
}

var menuItems = []menuItem{
	{actionList, "Notes", "Browse indexed notes"},
	{actionCapture, "Capture", "Create an inbox note and edit it in Neovim"},
	{actionPage, "Page", "Create a page, optionally inside an area"},
	{actionBookmark, "Bookmark", "Save a web address"},
	{actionJournal, "Journal", "Open today's journal editor"},
	{actionRebuild, "Rebuild index", "Index the AsciiDoc notes in the vault"},
	{actionPerson, "Person", "Create a person and organize them in groups"},
}
