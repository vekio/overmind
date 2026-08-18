package server

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
}

func (handler *rebuildIndexHandlerStub) Handle(
	context.Context,
	rebuildindex.RebuildIndexCommand,
) (rebuildindex.RebuildIndexResult, error) {
	handler.called = true
	return rebuildindex.RebuildIndexResult{Documents: 3}, nil
}

func TestIndexRebuildUsesServerApplication(t *testing.T) {
	handler := &rebuildIndexHandlerStub{}
	command := newTestCommand(t, func(config.ServerConfig) (Runtime, error) {
		return runtimeStub{application: &app.Application{
			Commands: app.Commands{RebuildIndex: handler},
		}}, nil
	})
	var output bytes.Buffer
	command.Writer = &output

	err := command.Run(context.Background(), []string{
		"overmind-server", "--config", writeServerConfig(t), "index", "rebuild",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !handler.called || output.String() != "Indexed 3 documents\n" {
		t.Fatalf("called = %t, output = %q", handler.called, output.String())
	}
}
