package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestListDocumentsWritesPipelineFriendlyTSV(t *testing.T) {
	configPath, _ := writeLocalConfig(t)
	create := newTestCommand(t)
	var created bytes.Buffer
	create.Writer = &created
	if err := create.Run(context.Background(), []string{
		"overmind", "--config", configPath,
		"page", "First", "--area", "Knowledge/Go", "--tag", "Go", "--tag", "DDD",
	}); err != nil {
		t.Fatalf("create page: %v", err)
	}

	list := newTestCommand(t)
	var output bytes.Buffer
	list.Writer = &output
	if err := list.Run(context.Background(), []string{
		"overmind", "--config", configPath,
		"ls", "--type", "page", "--title", "First", "--area", "Knowledge", "--tag", "Go", "--tag", "DDD",
	}); err != nil {
		t.Fatalf("list documents: %v", err)
	}

	id := strings.TrimSpace(created.String())
	want := id + "\tpage\tknowledge/go\tFirst\tgo,ddd\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestListDocumentsRejectsArguments(t *testing.T) {
	command := newTestCommand(t)
	err := command.Run(context.Background(), []string{"overmind", "ls", "page"})
	if err == nil || !strings.Contains(err.Error(), "unexpected arguments") {
		t.Fatalf("Run() error = %v", err)
	}
}
