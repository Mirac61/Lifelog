package ui

import (
	"fmt"
	"time"
)

// The German calendar, in one place: the pages and the view models that feed
// them both write dates, and two tables would drift apart.
var (
	MonthShort   = [...]string{"", "Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"}
	MonthLong    = [...]string{"", "Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"}
	WeekdayShort = [...]string{"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"}
	WeekdayLong  = [...]string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}
)

// DayMonth renders "26 Aug", the compact form the lists use.
func DayMonth(d time.Time) string { return fmt.Sprintf("%d %s", d.Day(), MonthShort[d.Month()]) }

// DayMonthYear renders "2. Jan 2026".
func DayMonthYear(d time.Time) string {
	return fmt.Sprintf("%d. %s %d", d.Day(), MonthShort[d.Month()], d.Year())
}

// LongDate renders "Sonntag, 6. September 2026 · KW 36" — the page head's line.
// The space before the week number is non-breaking, as in the design.
func LongDate(d time.Time) string {
	_, week := d.ISOWeek()
	return fmt.Sprintf("%s, %d. %s %d · KW %d", WeekdayLong[d.Weekday()], d.Day(), MonthLong[d.Month()], d.Year(), week)
}

func DaysInYear(year int) int {
	return time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC).YearDay()
}

// BestDaySuffix renders " · <date>" when a best date exists, or nothing.
func BestDaySuffix(date string) string {
	if date == "" {
		return ""
	}
	return " · " + FormatShortDate(date)
}

// FormatShortDate renders an ISO date as "2 Jan".
func FormatShortDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return DayMonth(t)
}

// Head is the sticky page head. ID is only needed where htmx swaps the whole
// header out of band.
type Head struct {
	ID    string
	Title string
	Sub   string
	OOB   bool // answer an htmx swap: replace the header with this one
}
