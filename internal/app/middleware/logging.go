// Package middleware contains decorators for application use-case handlers.
package middleware

import (
	"context"
	"log/slog"
	"time"
)

// Handler is the common shape of command and query handlers.
type Handler[Input, Output any] interface {
	Handle(context.Context, Input) (Output, error)
}

// Logging decorates a handler with debug-level completion and failure logs.
func Logging[Input, Output any](
	name string,
	next Handler[Input, Output],
	logger *slog.Logger,
) Handler[Input, Output] {
	if name == "" {
		panic("logging middleware requires use-case name")
	}
	if next == nil {
		panic("logging middleware requires next handler")
	}
	if logger == nil {
		panic("logging middleware requires logger")
	}
	return &loggingHandler[Input, Output]{name: name, next: next, logger: logger}
}

type loggingHandler[Input, Output any] struct {
	name   string
	next   Handler[Input, Output]
	logger *slog.Logger
}

func (handler *loggingHandler[Input, Output]) Handle(
	ctx context.Context,
	input Input,
) (Output, error) {
	startedAt := time.Now()
	result, err := handler.next.Handle(ctx, input)
	if err != nil {
		handler.logger.DebugContext(ctx,
			"use case failed",
			"use_case", handler.name,
			"duration", time.Since(startedAt),
			"error", err,
		)
		return result, err
	}

	handler.logger.DebugContext(ctx,
		"use case completed",
		"use_case", handler.name,
		"duration", time.Since(startedAt),
	)
	return result, nil
}
