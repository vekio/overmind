package httpclient

import "fmt"

// ResponseError reports an unsuccessful response from the remote API.
type ResponseError struct {
	StatusCode int
	Message    string
}

func (err *ResponseError) Error() string {
	return fmt.Sprintf("remote Overmind returned HTTP %d: %s", err.StatusCode, err.Message)
}
