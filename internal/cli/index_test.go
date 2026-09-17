package cli

import (
	"bytes"
	"context"
	"testing"
)

func TestIndexRebuildExecutesLocalUseCase(t *testing.T) {
	configPath, _ := writeLocalConfig(t)
	create := newTestCommand(t)
	create.Writer = &bytes.Buffer{}
	if err := create.Run(context.Background(), []string{
		"overmind", "--config", configPath, "page", "First page",
	}); err != nil {
		t.Fatalf("create page: %v", err)
	}

	rebuild := newTestCommand(t)
	var output bytes.Buffer
	rebuild.Writer = &output
	if err := rebuild.Run(context.Background(), []string{
		"overmind", "--config", configPath, "index", "rebuild",
	}); err != nil {
		t.Fatalf("rebuild index: %v", err)
	}
	if output.String() != "Indexed 1 documents\n" {
		t.Fatalf("output = %q", output.String())
	}
}
