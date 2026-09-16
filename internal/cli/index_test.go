package cli

import (
	"bytes"
	"context"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/app/rebuildindex"
	"git.casta.me/alberto/overmind/internal/config"
)

type rebuildIndexHandlerStub struct {
	called bool
	result rebuildindex.RebuildIndexResult
}

func (handler *rebuildIndexHandlerStub) Handle(context.Context, rebuildindex.RebuildIndexCommand) (rebuildindex.RebuildIndexResult, error) {
	handler.called = true
	return handler.result, nil
}

func TestIndexRebuildExecutesLocalUseCase(t *testing.T) {
	configPath := writeLocalConfig(t)
	handler := &rebuildIndexHandlerStub{result: rebuildindex.RebuildIndexResult{Documents: 2}}
	command := newTestCommand(t, func(config.Config) (Runtime, error) {
		return runtimeStub{application: &app.Application{
			Commands: app.Commands{RebuildIndex: handler},
		}}, nil
	})
	var output bytes.Buffer
	command.Writer = &output

	if err := command.Run(context.Background(), []string{"overmind", "--config", configPath, "index", "rebuild"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !handler.called || output.String() != "Indexed 2 documents\n" {
		t.Fatalf("called = %t, output = %q", handler.called, output.String())
	}
}
