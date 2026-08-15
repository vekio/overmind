package middleware

import (
	"bytes"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggerRecordsRequestAndResponse(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	handler := Logger(logger)(stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, _ *stdhttp.Request) {
		response.WriteHeader(stdhttp.StatusCreated)
		_, _ = response.Write([]byte("created"))
	}))
	request := httptest.NewRequest(stdhttp.MethodPost, "/pages", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	logEntry := output.String()
	for _, value := range []string{
		`"msg":"HTTP request"`,
		`"method":"POST"`,
		`"path":"/pages"`,
		`"status":201`,
		`"bytes":7`,
	} {
		if !strings.Contains(logEntry, value) {
			t.Fatalf("log entry = %q, want %q", logEntry, value)
		}
	}
}
