// Package response provides shared HTTP response writers.
package response

import (
	"encoding/json"
	stdhttp "net/http"
)

// JSON writes a JSON response with the supplied status.
func JSON(writer stdhttp.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

// Error writes a JSON error response.
func Error(writer stdhttp.ResponseWriter, status int, message string) {
	JSON(writer, status, struct {
		Error string `json:"error"`
	}{Error: message})
}
