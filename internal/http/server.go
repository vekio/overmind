// Package http implements Overmind's HTTP input adapter.
package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	stdhttp "net/http"
	"strconv"
	"time"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/http/middleware"
)

const shutdownTimeout = 10 * time.Second

// Server wraps the standard library HTTP server.
type Server struct {
	server *stdhttp.Server
}

// NewServer creates an HTTP server exposing application use cases.
func NewServer(address string, application *app.Application, logger *slog.Logger) (*Server, error) {
	if application == nil {
		panic("http server requires application")
	}
	if logger == nil {
		panic("http server requires logger")
	}
	if err := validateAddress(address); err != nil {
		return nil, err
	}

	httpLogger := logger.With("component", "http")
	return &Server{
		server: &stdhttp.Server{
			Addr:              address,
			Handler:           middleware.Logger(httpLogger)(routes(application, httpLogger)),
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}, nil
}

func validateAddress(address string) error {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid HTTP address %q: %w", address, err)
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return fmt.Errorf("invalid HTTP address port %q", port)
	}
	return nil
}

// Serve listens until the context is cancelled or the server fails. The
// optional onListening callback runs after the listener has opened.
func (server *Server) Serve(ctx context.Context, onListening func(string) error) error {
	if err := ctx.Err(); err != nil {
		return nil
	}
	listener, err := net.Listen("tcp", server.server.Addr)
	if err != nil {
		return fmt.Errorf("listen HTTP on %s: %w", server.server.Addr, err)
	}
	if onListening != nil {
		if err := onListening(listener.Addr().String()); err != nil {
			_ = listener.Close()
			return fmt.Errorf("announce HTTP server: %w", err)
		}
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.server.Serve(listener)
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, stdhttp.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP on %s: %w", server.server.Addr, err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		if err := <-errCh; err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
			return fmt.Errorf("serve HTTP on %s: %w", server.server.Addr, err)
		}
		return nil
	}
}
