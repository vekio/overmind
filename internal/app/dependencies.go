package app

import "git.casta.me/alberto/overmind/internal/ports"

// Dependencies contains the ports required by the application.
type Dependencies struct {
	IDGenerator ports.IDGenerator
	Renderer    ports.NoteRenderer
	Writer      ports.NoteWriter
	Reader      ports.NoteReader
	Index       ports.NoteIndex
	Lister      ports.NoteLister
	Walker      ports.NoteWalker
	Parser      ports.NoteParser
}
