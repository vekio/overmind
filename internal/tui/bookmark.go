package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/bookmark"
	"github.com/vekio/overmind/internal/domain/bookmarks"
)

// CreateBookmarkFunc connects form submissions to the bookmark creation use case.
type CreateBookmarkFunc func(context.Context, bookmark.CreateCommand) (bookmark.CreateResult, error)

func newBookmarkForm() form {
	return newForm("Bookmark",
		fieldSpec{id: "url", label: "URL", kind: fieldText, placeholder: "Required, https://example.com", validate: func(raw string) error {
			_, err := bookmarks.NewURL(raw)
			return err
		}},
		contentField("Write your bookmark…"),
		tagsField(),
	)
}

func (m model) saveBookmark(values map[string]string) tea.Cmd {
	command := bookmark.CreateCommand{URL: values["url"], Content: values["content"], Tags: formListValues(values["tags"])}
	return func() tea.Msg {
		if m.createBookmark == nil {
			return noteCreated{action: actionBookmark, err: fmt.Errorf("bookmark creation is not configured")}
		}
		_, err := m.createBookmark(m.ctx, command)
		return noteCreated{action: actionBookmark, err: err}
	}
}
