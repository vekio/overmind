package createpage

import "git.casta.me/alberto/overmind/internal/domain"

// CreatePageCommand is the input shared by CLI, HTTP, and future input
// adapters.
type CreatePageCommand struct {
	Title string
	Area  string
}

// CreatePageResult is the minimal output of creating a page.
type CreatePageResult struct {
	ID domain.DocumentID
}
