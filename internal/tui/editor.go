package tui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"uuid"

	tea "charm.land/bubbletea/v2"
)

type noteEditorState struct {
	id               uuid.UUID
	original         []byte
	draftPath        string
	createdPath      string
	returnScreen     screen
	unchangedMessage string
}

type noteOpenResult struct {
	id               uuid.UUID
	source           []byte
	createdPath      string
	returnScreen     screen
	unchangedMessage string
	err              error
}

type noteEditorFinished struct {
	err error
}

type noteSaveResult struct {
	err error
}

func (m model) openNoteEditor(message noteOpenResult) (tea.Model, tea.Cmd) {
	if message.err != nil {
		return m.noteEditorOpenError(message, message.err)
	}
	if _, err := exec.LookPath("nvim"); err != nil {
		return m.noteEditorOpenError(message, fmt.Errorf("nvim was not found in PATH"))
	}
	draftPath, err := createNoteDraft(message.source)
	if err != nil {
		return m.noteEditorOpenError(message, err)
	}
	m.noteEditor = noteEditorState{
		id: message.id, original: bytes.Clone(message.source), draftPath: draftPath,
		createdPath: message.createdPath, returnScreen: message.returnScreen,
		unchangedMessage: message.unchangedMessage,
	}
	m.problem = ""
	m.screen = screenBusy
	command := exec.CommandContext(m.ctx, "nvim", draftPath)
	return m, tea.ExecProcess(command, func(err error) tea.Msg {
		return noteEditorFinished{err: err}
	})
}

func (m model) noteEditorOpenError(message noteOpenResult, err error) (tea.Model, tea.Cmd) {
	m.screen = message.returnScreen
	m.problem = fmt.Sprintf("open note: %v", err)
	if message.createdPath != "" {
		m.problem = fmt.Sprintf("note created at %s; %s", message.createdPath, m.problem)
	}
	return m, nil
}

func (m model) finishNoteEdit(editorErr error) (tea.Model, tea.Cmd) {
	m.screen = m.noteEditor.returnScreen
	if editorErr != nil {
		m.problem = fmt.Sprintf("Neovim exited with an error; draft kept at %s: %v", m.noteEditor.draftPath, editorErr)
		return m, nil
	}
	source, err := os.ReadFile(m.noteEditor.draftPath)
	if err != nil {
		m.problem = fmt.Sprintf("read edited note; draft kept at %s: %v", m.noteEditor.draftPath, err)
		return m, nil
	}
	if bytes.Equal(source, m.noteEditor.original) {
		removeDraft(m.noteEditor.draftPath)
		message := m.noteEditor.unchangedMessage
		kind := notificationInfo
		if m.noteEditor.createdPath != "" {
			kind = notificationSuccess
		}
		m.noteEditor = noteEditorState{}
		m.problem = ""
		return m, m.notify(message, kind)
	}
	m.screen = screenBusy
	id := m.noteEditor.id
	original := bytes.Clone(m.noteEditor.original)
	return m, func() tea.Msg {
		_, err := m.client.UpdateNote(m.ctx, id, original, source)
		return noteSaveResult{err: err}
	}
}

func (m model) finishNoteSave(message noteSaveResult) (tea.Model, tea.Cmd) {
	returnScreen := m.noteEditor.returnScreen
	m.screen = returnScreen
	if message.err != nil {
		m.problem = fmt.Sprintf("save note; edited draft kept at %s: %v", m.noteEditor.draftPath, message.err)
		return m, nil
	}
	removeDraft(m.noteEditor.draftPath)
	m.noteEditor = noteEditorState{}
	m.problem = ""
	notificationCmd := m.notify("Note saved", notificationSuccess)
	if returnScreen == screenNotes {
		next, loadCmd := m.loadNotes()
		return next, tea.Batch(notificationCmd, loadCmd)
	}
	return m, notificationCmd
}

func createNoteDraft(source []byte) (string, error) {
	file, err := os.CreateTemp("", "overmind-note-*.adoc")
	if err != nil {
		return "", fmt.Errorf("create note draft: %w", err)
	}
	path := file.Name()
	if _, err := file.Write(source); err != nil {
		_ = file.Close()
		removeDraft(path)
		return "", fmt.Errorf("write note draft: %w", err)
	}
	if err := file.Close(); err != nil {
		removeDraft(path)
		return "", fmt.Errorf("close note draft: %w", err)
	}
	return path, nil
}

func removeDraft(path string) {
	if path != "" {
		_ = os.Remove(path)
	}
}
