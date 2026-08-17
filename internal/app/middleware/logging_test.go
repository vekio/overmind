package middleware

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

type handlerStub struct {
	result string
	err    error
}

func (handler handlerStub) Handle(context.Context, int) (string, error) {
	return handler.result, handler.err
}

func TestLoggingPreservesResultAndLogsCompletion(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	handler := Logging[int, string]("create_page", handlerStub{result: "created"}, logger)

	result, err := handler.Handle(context.Background(), 1)
	if err != nil || result != "created" {
		t.Fatalf("Handle() = (%q, %v)", result, err)
	}
	if log := output.String(); !strings.Contains(log, `"msg":"use case completed"`) ||
		!strings.Contains(log, `"use_case":"create_page"`) ||
		!strings.Contains(log, `"duration":`) {
		t.Fatalf("log = %s", log)
	}
}

func TestLoggingPreservesAndLogsFailure(t *testing.T) {
	operationErr := errors.New("template unavailable")
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	handler := Logging[int, string]("create_page", handlerStub{err: operationErr}, logger)

	_, err := handler.Handle(context.Background(), 1)
	if !errors.Is(err, operationErr) {
		t.Fatalf("Handle() error = %v", err)
	}
	if log := output.String(); !strings.Contains(log, `"msg":"use case failed"`) ||
		!strings.Contains(log, `"use_case":"create_page"`) ||
		!strings.Contains(log, `"error":"template unavailable"`) {
		t.Fatalf("log = %s", log)
	}
}
