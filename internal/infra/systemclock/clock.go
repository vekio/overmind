// Package systemclock implements a clock backed by the system time.
package systemclock

import (
	"time"

	"git.casta.me/alberto/overmind/internal/ports"
)

var _ ports.Clock = Clock{}

// Clock provides the current system time.
type Clock struct{}

// New creates a system clock.
func New() Clock { return Clock{} }

// Now returns the current time in UTC.
func (Clock) Now() time.Time { return time.Now().UTC() }
