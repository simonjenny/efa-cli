package carbon

import (
	"testing"
	"time"
)

func mustParse(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := Parse(s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return ts
}

func TestDiffForHumansGoldenVectors(t *testing.T) {
	now := mustParse(t, "2026-09-14T20:41:10Z")
	cases := []struct {
		in   string
		want string
	}{
		{"2026-09-14T20:41:10Z", "0 seconds ago"},
		{"2026-09-14T20:41:40Z", "30 seconds from now"},
		{"2026-09-14T20:40:25Z", "45 seconds ago"},
		{"2026-09-14T20:42:10Z", "1 minute from now"},
		{"2026-09-14T20:49:00Z", "7 minutes from now"},
		{"2026-09-14T22:41:10Z", "2 hours from now"},
		{"2026-09-14T20:30:00Z", "11 minutes ago"},
		{"2026-09-13T19:41:10Z", "1 day ago"},
		{"2026-09-17T20:41:10Z", "3 days from now"},
		{"2026-09-22T20:41:10Z", "1 week from now"},
		{"2026-10-19T20:41:10Z", "1 month from now"},
		{"2026-12-14T20:41:10Z", "3 months from now"},
		{"2027-09-14T20:41:10Z", "1 year from now"},
	}
	for _, c := range cases {
		got := DiffForHumans(mustParse(t, c.in), now)
		if got != c.want {
			t.Errorf("DiffForHumans(%s) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDiffInMinutesGoldenVectors(t *testing.T) {
	cases := []struct {
		a, b string
		want float64
	}{
		{"20:42", "20:53", 11.0},
		{"20:42:00", "20:53:30", 11.5},
		{"20:42", "20:41", -1.0},
		{"23:59", "00:01", -1438.0},
	}
	for _, c := range cases {
		got := DiffInMinutes(mustParse(t, c.a), mustParse(t, c.b))
		if got != c.want {
			t.Errorf("DiffInMinutes(%s, %s) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestFloorMinutesGoldenVectors(t *testing.T) {
	now := mustParse(t, "2026-09-14T20:41:10Z")
	cases := []struct {
		in   string
		want string
	}{
		{"2026-09-14T20:49:00Z", "7"},
		{"2026-09-14T20:30:00Z", "-12"},
	}
	for _, c := range cases {
		got := FloorMinutes(DiffInMinutes(now, mustParse(t, c.in)))
		if got != c.want {
			t.Errorf("FloorMinutes(%s) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDurationString(t *testing.T) {
	cases := []struct {
		a, b string
		want string
	}{
		{"20:42", "20:53", "11"},
		{"20:42:00", "20:53:30", "11.5"},
		{"20:53", "20:42", "-11"},
	}
	for _, c := range cases {
		got := FloatString(DiffInMinutes(mustParse(t, c.a), mustParse(t, c.b)))
		if got != c.want {
			t.Errorf("duration %s->%s = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}

func TestParseFormats(t *testing.T) {
	ts := mustParse(t, "2026-09-14T20:43:00Z")
	if ts.Location() != time.UTC {
		t.Errorf("expected UTC location")
	}
	if ts.Hour() != 20 || ts.Minute() != 43 {
		t.Errorf("wrong time: %v", ts)
	}
	hm := mustParse(t, "20:42")
	now := time.Now()
	if hm.Year() != now.Year() || hm.Month() != now.Month() || hm.Day() != now.Day() {
		t.Errorf("HH:MM should refer to today: %v", hm)
	}
	if _, err := Parse("not a time"); err == nil {
		t.Error("expected parse error")
	}
}
