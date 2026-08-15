// Package middleware provides HTTP middleware shared by all routes.
package middleware

import (
	"log/slog"
	stdhttp "net/http"
	"time"
)

type loggingResponseWriter struct {
	stdhttp.ResponseWriter
	status int
	bytes  int
}

func (response *loggingResponseWriter) WriteHeader(status int) {
	if response.status != 0 {
		return
	}
	response.status = status
	response.ResponseWriter.WriteHeader(status)
}

func (response *loggingResponseWriter) Write(content []byte) (int, error) {
	if response.status == 0 {
		response.WriteHeader(stdhttp.StatusOK)
	}
	written, err := response.ResponseWriter.Write(content)
	response.bytes += written
	return written, err
}

// Unwrap allows http.ResponseController to access the original writer.
func (response *loggingResponseWriter) Unwrap() stdhttp.ResponseWriter {
	return response.ResponseWriter
}

// Logger records request and response details.
func Logger(logger *slog.Logger) func(stdhttp.Handler) stdhttp.Handler {
	if logger == nil {
		panic("HTTP logger middleware requires logger")
	}

	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(writer stdhttp.ResponseWriter, request *stdhttp.Request) {
			startedAt := time.Now()
			response := &loggingResponseWriter{ResponseWriter: writer}

			defer func() {
				status := response.status
				if status == 0 {
					status = stdhttp.StatusOK
				}
				logger.InfoContext(request.Context(), "HTTP request",
					"method", request.Method,
					"path", request.URL.Path,
					"status", status,
					"bytes", response.bytes,
					"duration", time.Since(startedAt),
					"remote_address", request.RemoteAddr,
				)
			}()

			next.ServeHTTP(response, request)
		})
	}
}
