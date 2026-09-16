package createpage

import "git.casta.me/alberto/overmind/internal/domain"

// CreatePageCommand is the input of creating a page.
type CreatePageCommand struct {
	Title string
	Area  string
	Tags  []string
}

// CreatePageResult is the minimal output of creating a page.
type CreatePageResult struct {
	ID   domain.DocumentID
	Path string
}
