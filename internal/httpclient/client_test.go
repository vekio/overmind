package httpclient

import (
	stdhttp "net/http"
	"testing"
)

func TestNewRejectsInvalidEndpoint(t *testing.T) {
	for _, endpoint := range []string{
		"overmind.example",
		"ftp://overmind.example",
		"https://overmind.example?token=secret",
		"https://overmind.example#fragment",
	} {
		t.Run(endpoint, func(t *testing.T) {
			if _, err := New(endpoint, &stdhttp.Client{}); err == nil {
				t.Fatal("New() error = nil")
			}
		})
	}
}
