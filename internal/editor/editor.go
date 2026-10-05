// Package editor opens private note drafts in the user's external editor.
package editor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/vekio/x/file"
)

// Draft owns a temporary AsciiDoc file. It never writes into the vault.
type Draft struct {
	Path     string
	original []byte
	keep     bool
}

// New copies source into a private temporary .adoc file and retains the initial snapshot.
func New(source []byte) (*Draft, error) {
	file, err := os.CreateTemp("", "overmind-*.adoc")
	if err != nil {
		return nil, fmt.Errorf("create editor draft: %w", err)
	}
	path := file.Name()
	_, writeErr := file.Write(source)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(path)
		if writeErr != nil {
			return nil, writeErr
		}
		return nil, closeErr
	}
	return &Draft{Path: path, original: append([]byte(nil), source...)}, nil
}

// Command resolves EDITOR, then VISUAL, then vi, supporting shell arguments.
// The draft path is passed separately so spaces and shell syntax in it remain literal.
func (draft *Draft) Command(ctx context.Context) *exec.Cmd {
	command := strings.TrimSpace(os.Getenv("EDITOR"))
	if command == "" {
		command = strings.TrimSpace(os.Getenv("VISUAL"))
	}
	if command == "" {
		command = "vi"
	}
	return exec.CommandContext(ctx, "sh", "-c", "exec "+command+` "$@"`, "overmind-editor", draft.Path)
}

// Read returns the current draft bytes after the external editor has exited.
func (draft *Draft) Read() ([]byte, error) {
	source, err := os.ReadFile(draft.Path)
	if err != nil {
		return nil, fmt.Errorf("read editor draft: %w", err)
	}
	return source, nil
}

// Close discards the draft file, including drafts explicitly marked for retention.
func (draft *Draft) Close() {
	if draft != nil {
		_ = os.Remove(draft.Path)
	}
}

// Write refreshes the private draft after managed metadata has been saved.
func (draft *Draft) Write(source []byte) error {
	return file.WriteAtomic(draft.Path, source, 0600)
}

// Keep preserves even an unchanged draft after a failed edit or explicit request.
func (draft *Draft) Keep() {
	if draft != nil {
		draft.keep = true
	}
}

// Finish removes unchanged drafts and retains unfinished work when the UI exits.
func (draft *Draft) Finish() (string, error) {
	if draft == nil {
		return "", nil
	}
	source, err := draft.Read()
	if err != nil {
		return draft.Path, err
	}
	if draft.keep || !bytes.Equal(source, draft.original) {
		return draft.Path, nil
	}
	draft.Close()
	return "", nil
}
