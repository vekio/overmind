package app

import (
	"context"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

// CaptureCommand contains the input for capturing an inbox note.
type CaptureCommand struct {
	Content string
}

// CaptureResult contains the captured inbox note and its path.
type CaptureResult struct {
	Inbox domain.Inbox
	Path  string
}

// CaptureHandler captures inbox notes.
type CaptureHandler struct {
	saver *noteSaver
	ids   ports.IDGenerator
}

func newCaptureHandler(saver *noteSaver, ids ports.IDGenerator) *CaptureHandler {
	return &CaptureHandler{saver: saver, ids: ids}
}

// Handle creates, persists and indexes an inbox note.
func (handler *CaptureHandler) Handle(ctx context.Context, command CaptureCommand) (CaptureResult, error) {
	inbox, err := domain.NewInbox(handler.ids.Generate(), command.Content, time.Now())
	if err != nil {
		return CaptureResult{}, err
	}
	path, err := handler.saver.save(ctx, inbox)
	if err != nil {
		return CaptureResult{}, err
	}

	return CaptureResult{Inbox: inbox, Path: path}, nil
}
