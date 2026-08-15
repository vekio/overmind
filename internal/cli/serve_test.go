package cli

import "testing"

func TestLocalHTTPURL(t *testing.T) {
	for name, test := range map[string]struct {
		address string
		want    string
	}{
		"loopback":       {address: "127.0.0.1:8080", want: "http://127.0.0.1:8080/"},
		"all interfaces": {address: "0.0.0.0:8080", want: "http://localhost:8080/"},
		"IPv6":           {address: "[::1]:8080", want: "http://[::1]:8080/"},
	} {
		t.Run(name, func(t *testing.T) {
			if got := localHTTPURL(test.address); got != test.want {
				t.Fatalf("localHTTPURL(%q) = %q, want %q", test.address, got, test.want)
			}
		})
	}
}
