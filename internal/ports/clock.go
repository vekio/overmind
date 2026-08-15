// Package ports defines the boundaries between the application and its
// infrastructure adapters.
package ports

import "time"

// Clock provides the current time.
type Clock interface {
	Now() time.Time
}
