package calendar_test

import (
	"testing"

	"github.com/vekio/overmind/internal/domain/calendar"
)

func dateForTest(t *testing.T, value string) calendar.Date {
	t.Helper()
	date, err := calendar.NewDate(value)
	if err != nil {
		t.Fatal(err)
	}
	return date
}

func TestPeriodWindowsRespectCalendarBoundaries(t *testing.T) {
	cases := []struct {
		period           calendar.Period
		date, start, end string
	}{
		{calendar.Day, "2026-10-03", "2026-10-03", "2026-10-04"},
		{calendar.Day, "2024-02-29", "2024-02-29", "2024-03-01"},
		{calendar.Day, "2026-12-31", "2026-12-31", "2027-01-01"},
		{calendar.Week, "2026-10-03", "2026-09-28", "2026-10-05"},
		{calendar.Week, "2026-10-04", "2026-09-28", "2026-10-05"},
		{calendar.Week, "2026-10-05", "2026-10-05", "2026-10-12"},
		{calendar.Week, "2027-01-01", "2026-12-28", "2027-01-04"},
		{calendar.Month, "2026-10-03", "2026-10-01", "2026-11-01"},
		{calendar.Month, "2024-02-29", "2024-02-01", "2024-03-01"},
		{calendar.Month, "2025-02-28", "2025-02-01", "2025-03-01"},
		{calendar.Month, "2026-12-31", "2026-12-01", "2027-01-01"},
		{calendar.Day, "0001-01-01", "0001-01-01", "0001-01-02"},
	}
	for _, test := range cases {
		t.Run(test.period.String()+"/"+test.date, func(t *testing.T) {
			date := dateForTest(t, test.date)
			window, err := test.period.Window(date)
			if err != nil {
				t.Fatal(err)
			}
			if !window.Start().Equal(dateForTest(t, test.start)) || !window.End().Equal(dateForTest(t, test.end)) {
				t.Fatalf("window = [%s, %s), want [%s, %s)", window.Start(), window.End(), test.start, test.end)
			}
			if !window.Contains(date) || !window.Contains(window.Start()) || window.Contains(window.End()) {
				t.Fatal("window must contain the supplied date and start, but exclude end")
			}
			next, err := test.period.Window(window.End())
			if err != nil || !next.Start().Equal(window.End()) || next.Contains(window.Start()) {
				t.Fatalf("adjacent window should start at the excluded end: %v", err)
			}
		})
	}
}

func TestWindowRejectsInvalidBoundsAndDates(t *testing.T) {
	start := dateForTest(t, "2026-10-01")
	end := dateForTest(t, "2026-11-01")
	for _, bounds := range [][2]calendar.Date{{start, start}, {end, start}, {{}, end}, {start, {}}} {
		if _, err := calendar.NewWindow(bounds[0], bounds[1]); err == nil {
			t.Fatalf("invalid bounds accepted: %v", bounds)
		}
	}
	window, err := calendar.NewWindow(start, end)
	if err != nil {
		t.Fatal(err)
	}
	if window.Contains(calendar.Date{}) || window.Contains(dateForTest(t, "2026-09-30")) || window.Contains(dateForTest(t, "2026-11-02")) {
		t.Fatal("uninitialized and out-of-range dates must not belong to a window")
	}
	if (calendar.Window{}).Contains(start) {
		t.Fatal("an uninitialized window must not contain dates")
	}
	for _, period := range []calendar.Period{calendar.Day, calendar.Week, calendar.Month, calendar.Period(255), calendar.Unknown} {
		if _, err := period.Window(calendar.Date{}); err == nil {
			t.Fatalf("zero date accepted by %q", period)
		}
	}
	for _, period := range []calendar.Period{calendar.Unknown, calendar.Period(255)} {
		if _, err := period.Window(start); err == nil {
			t.Fatal("unsupported period accepted")
		}
	}
	if !(calendar.Date{}).IsZero() {
		t.Fatal("uninitialized date must be zero")
	}
}
