package http

import (
	"io"
	"log/slog"
	"testing"
)

func TestNewServerConfiguresStandardServer(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server, err := NewServer("127.0.0.1:8080", logger)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server.server.Addr != "127.0.0.1:8080" || server.server.Handler == nil {
		t.Fatalf("server = %#v", server.server)
	}
}

func TestNewServerRejectsInvalidAddress(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if _, err := NewServer("localhost", logger); err == nil {
		t.Fatal("NewServer() error = nil")
	}
}
