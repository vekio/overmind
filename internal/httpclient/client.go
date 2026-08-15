// Package httpclient implements access to a remote Overmind HTTP API.
package httpclient

import (
	"context"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"strings"
)

// Client sends requests relative to an Overmind API endpoint.
type Client struct {
	http    *stdhttp.Client
	baseURL *url.URL
}

// New creates an Overmind HTTP client.
func New(endpoint string, client *stdhttp.Client) (*Client, error) {
	if client == nil {
		panic("Overmind HTTP client requires standard HTTP client")
	}

	baseURL, err := url.Parse(endpoint)
	if err != nil || (baseURL.Scheme != "http" && baseURL.Scheme != "https") || baseURL.Host == "" {
		return nil, fmt.Errorf("invalid Overmind endpoint %q", endpoint)
	}
	if baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return nil, fmt.Errorf("Overmind endpoint must not contain query or fragment")
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + "/"

	return &Client{http: client, baseURL: baseURL}, nil
}

// NewRequest creates a request relative to the configured endpoint.
func (client *Client) NewRequest(ctx context.Context, method, relativePath string, body io.Reader) (*stdhttp.Request, error) {
	reference, err := url.Parse(relativePath)
	if err != nil {
		return nil, fmt.Errorf("parse API path %q: %w", relativePath, err)
	}
	request, err := stdhttp.NewRequestWithContext(ctx, method, client.baseURL.ResolveReference(reference).String(), body)
	if err != nil {
		return nil, fmt.Errorf("create HTTP request: %w", err)
	}
	return request, nil
}

// Do sends an HTTP request.
func (client *Client) Do(request *stdhttp.Request) (*stdhttp.Response, error) {
	return client.http.Do(request)
}
