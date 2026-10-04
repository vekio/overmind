package tui

type action uint8

const (
	actionList action = iota
	actionInbox
	actionHabit
	actionPerson
	actionBookmark
	actionPage
	actionJournal
	actionReindex
)

func (a action) String() string {
	for _, item := range menuItems {
		if item.action == a {
			return item.title
		}
	}
	return "Unknown"
}

type menuItem struct {
	action             action
	title, description string
}

var menuItems = []menuItem{
	{actionList, "Notes", "Browse and filter saved notes"},
	{actionInbox, "Inbox", "Create a quick note"},
	{actionHabit, "Habit", "Define a habit"},
	{actionPerson, "Person", "Create a person note"},
	{actionBookmark, "Bookmark", "Save a web address"},
	{actionPage, "Page", "Create a titled note"},
	{actionJournal, "Journal", "Create a daily entry"},
	{actionReindex, "Reindex", "Rebuild the index from note files"},
}
