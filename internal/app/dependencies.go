package app

import "git.casta.me/alberto/overmind/internal/ports"

// Dependencies contains the ports required by the application.
type Dependencies struct {
	IDGenerator ports.IDGenerator
	Renderer    ports.NoteRenderer
	Writer      ports.NoteWriter
	Index       ports.NoteIndex
}
