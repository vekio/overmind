package tui

type action uint8

const (
	actionCapture action = iota
	actionPage
	actionBookmark
	actionJournal
	actionRebuild
	actionList
)

func (a action) String() string {
	switch a {
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
	{actionCapture, "Capture", "Save a quick inbox note"},
	{actionPage, "Page", "Create a page, optionally inside an area"},
	{actionBookmark, "Bookmark", "Save a web address"},
	{actionJournal, "Journal", "Create today's journal note"},
	{actionRebuild, "Rebuild index", "Index the AsciiDoc notes in the vault"},
}
