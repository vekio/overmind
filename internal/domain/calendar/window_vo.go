package calendar

import "fmt"

// Window is a value object representing a nonempty calendar interval [Start, End): start is included,
// end is excluded, so adjacent periods do not share dates.
type Window struct {
	start Date
	end   Date
}

func NewWindow(start, end Date) (Window, error) {
	if start.IsZero() || end.IsZero() || !start.Before(end) {
		return Window{}, fmt.Errorf("window requires initialized dates with start before end")
	}
	return Window{start: start, end: end}, nil
}

func (window Window) Start() Date { return window.start }
func (window Window) End() Date   { return window.end }

// Contains reports whether a date is within the interval. An uninitialized
// window or date never contains or belongs to a calendar interval.
func (window Window) Contains(date Date) bool {
	return !date.IsZero() && !window.start.IsZero() && !window.end.IsZero() &&
		!date.Before(window.start) && date.Before(window.end)
}
