package editor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditorArgumentsAndPrivateDraft(t *testing.T) {
	script := filepath.Join(t.TempDir(), "fake editor")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf '%s' \"$1\" > \"$2\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", "'"+script+"' 'argument with spaces'")
	draft, err := New([]byte("original"))
	if err != nil {
		t.Fatal(err)
	}
	defer draft.Close()
	info, err := os.Stat(draft.Path)
	if err != nil || info.Mode().Perm() != 0600 || !strings.HasSuffix(draft.Path, ".adoc") {
		t.Fatal("draft must be private AsciiDoc")
	}
	if err := draft.Command(context.Background()).Run(); err != nil {
		t.Fatal(err)
	}
	source, err := draft.Read()
	if err != nil || string(source) != "argument with spaces" {
		t.Fatalf("editor arguments: %q, %v", source, err)
	}
	draft.Close()
	if _, err := os.Stat(draft.Path); !os.IsNotExist(err) {
		t.Fatal("temporary file not removed")
	}
}

func TestEditorFallbacks(t *testing.T) {
	draft := &Draft{Path: "/tmp/note with spaces.adoc"}
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "visual-editor --wait")
	command := draft.Command(context.Background())
	if command.Args[2] != `exec visual-editor --wait "$@"` || command.Args[4] != draft.Path {
		t.Fatal(command.Args)
	}
	t.Setenv("VISUAL", "")
	if draft.Command(context.Background()).Args[2] != `exec vi "$@"` {
		t.Fatal("missing fallback")
	}
}

func TestDraftFinishPreservesUnfinishedWork(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		edit, keep, retained bool
	}{
		{"unchanged", false, false, false}, {"edited", true, false, true}, {"explicit keep", false, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			draft, err := New([]byte("original"))
			if err != nil {
				t.Fatal(err)
			}
			defer draft.Close()
			if tc.edit {
				if err := draft.Write([]byte("unfinished edit")); err != nil {
					t.Fatal(err)
				}
			}
			if tc.keep {
				draft.Keep()
			}
			path, err := draft.Finish()
			if err != nil || (path != "") != tc.retained {
				t.Fatalf("cleanup result: %q, %v", path, err)
			}
			_, statErr := os.Stat(draft.Path)
			if tc.retained && statErr != nil {
				t.Fatal("unfinished draft deleted")
			}
			if !tc.retained && !os.IsNotExist(statErr) {
				t.Fatal("unchanged draft not cleaned")
			}
		})
	}
}
