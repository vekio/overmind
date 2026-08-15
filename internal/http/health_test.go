package http

import (
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(stdhttp.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	handleHealth(response, request)

	if response.Code != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusOK)
	}
	if got, want := response.Body.String(), "{\"status\":\"ok\"}\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}
