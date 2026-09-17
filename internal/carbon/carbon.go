// Package carbon replicates the Carbon 3.8.2 behaviour used by the
// original efa-cli application: diffForHumans() and diffInMinutes().
package carbon

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/simonjenny/efa-cli/internal/i18n"
	"github.com/simonjenny/efa-cli/internal/jsonx"
)

// testNow overrides Now() (like Carbon::setTestNow).
var testNow *time.Time

// SetTestNow fixes the current time for tests.
func SetTestNow(t time.Time) {
	testNow = &t
}

// Now returns the current time (or the test time when set).
func Now() time.Time {
	if testNow != nil {
		return *testNow
	}
	return time.Now()
}

// Parse parses a date string the way Carbon::parse does for the formats
// used by efa-cli: RFC 3339 timestamps, bare "d.m.Y" dates and bare
// "HH:MM" / "HH:MM:SS" times (which refer to today).
func Parse(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("02.01.2006", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("15:04:05", s); err == nil {
		now := Now()
		return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, now.Location()), nil
	}
	if t, err := time.Parse("15:04", s); err == nil {
		now := Now()
		return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location()), nil
	}
	return time.Time{}, fmt.Errorf("unable to parse %q", s)
}

// DiffInMinutes returns the signed difference (date - this) in minutes as
// a float, matching Carbon 3.8.2's diffInMinutes($date, $absolute = false).
func DiffInMinutes(this, date time.Time) float64 {
	return date.Sub(this).Seconds() / 60
}

// DiffForHumans formats the difference between t and now the way
// Carbon::diffForHumans() does: the largest non-zero unit (year, month,
// week, day, hour, minute, second) with singular/plural and an "ago" /
// "from now" suffix.
func DiffForHumans(t, now time.Time) string {
	invert := t.After(now)
	earlier, later := t, now
	if invert {
		earlier, later = now, t
	}
	y, m, d, h, mi, s := intervalComponents(earlier, later)

	count, unit := 0, "second"
	switch {
	case y > 0:
		count, unit = y, "year"
	case m > 0:
		count, unit = m, "month"
	case d/7 > 0:
		count, unit = d/7, "week"
	case d%7 > 0:
		count, unit = d%7, "day"
	case h > 0:
		count, unit = h, "hour"
	case mi > 0:
		count, unit = mi, "minute"
	case s > 0:
		count, unit = s, "second"
	}

	unitKey := "diff." + unit
	if count != 1 {
		unitKey += "s"
	}
	text := fmt.Sprintf("%d %s", count, i18n.T(unitKey))
	if invert {
		return fmt.Sprintf(i18n.T("diff.from_now"), text)
	}
	return fmt.Sprintf(i18n.T("diff.ago"), text)
}

// intervalComponents computes the calendar difference t2 - t1 (t2 >= t1)
// like PHP's DateTime::diff in UTC.
func intervalComponents(t1, t2 time.Time) (y, m, d, h, mi, s int) {
	t1 = t1.UTC()
	t2 = t2.UTC()
	y1, mo1, d1 := t1.Date()
	y2, mo2, d2 := t2.Date()
	h1, mi1, s1 := t1.Clock()
	h2, mi2, s2 := t2.Clock()

	y = y2 - y1
	m = int(mo2) - int(mo1)
	d = d2 - d1
	h = h2 - h1
	mi = mi2 - mi1
	s = s2 - s1

	if s < 0 {
		s += 60
		mi--
	}
	if mi < 0 {
		mi += 60
		h--
	}
	if h < 0 {
		h += 24
		d--
	}
	if d < 0 {
		d += daysInMonth(y1, mo1)
		m--
	}
	if m < 0 {
		m += 12
		y--
	}
	return y, m, d, h, mi, s
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// FloorMinutes formats the signed minute difference as PHP would print a
// float: floor() followed by float-to-string conversion.
func FloorMinutes(minutes float64) string {
	return jsonx.FloatToPHPString(math.Floor(minutes))
}

// FloatString formats a float the way PHP converts a float to a string
// (serialize_precision=-1).
func FloatString(f float64) string {
	return jsonx.FloatToPHPString(f)
}

// FormatDateYmd formats a time as Ymd (PHP date('Ymd')).
func FormatDateYmd(t time.Time) string {
	return t.Format("20060102")
}

// FormatDateDMY formats a time as d.m.Y (PHP date('d.m.Y')).
func FormatDateDMY(t time.Time) string {
	return t.Format("02.01.2006")
}

// FormatTimeHi formats a time as H:i (PHP date('H:i')).
func FormatTimeHi(t time.Time) string {
	return t.Format("15:04")
}

// FormatTimeHiNumeric formats a time as Hi (PHP date('Hi')), matching the
// itdTime query parameter the EFA API expects.
func FormatTimeHiNumeric(t time.Time) string {
	return t.Format("1504")
}

// FormatInt converts an integer the way PHP prints it.
func FormatInt(i int) string {
	return strconv.Itoa(i)
}
