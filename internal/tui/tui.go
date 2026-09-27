// Package tui provides Overmind's interactive terminal interface.
package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
)

// Client provides the operations available in the terminal interface.
type Client interface {
	CreateBookmark(context.Context, domain.URL, domain.Tags) (string, error)
	Capture(context.Context, string) (string, error)
	CreateJournal(context.Context, domain.Tags) (string, error)
	CreatePage(context.Context, domain.Title, domain.Area, domain.Tags) (string, error)
	RebuildIndex(context.Context) (int, error)
	ListNotes(context.Context) ([]app.ListedNote, error)
}

// Run opens the terminal interface until the user quits.
func Run(ctx context.Context, client Client) error {
	_, err := tea.NewProgram(newModel(ctx, client), tea.WithContext(ctx)).Run()
	return err
}
