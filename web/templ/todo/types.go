package todo

import (
	"strings"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
)

type DayPage struct {
	Day      time.Time
	Todos    []store.Todo
	Upcoming []store.Todo // due after Day, earliest first
	Open     int          // unfinished on Day
	Planned  int          // their estimates, in minutes
}

// Preview is built from a store.Todo, so the capture bar cannot promise
// something the create handler would not save.
type Preview struct {
	Empty    bool
	Title    string
	Estimate string
	Due      string
	Category string
	Upcoming bool // due past the shown day: the row lands in "Kommende"
}

// A long title would push the chips behind the button.
const previewTitleMax = 42

// NewPreview formats a parsed todo. view is the shown day, today the real
// date — both ISO.
func NewPreview(t store.Todo, view, today string) Preview {
	p := Preview{Title: t.Text}
	if len([]rune(p.Title)) > previewTitleMax {
		p.Title = strings.TrimSpace(string([]rune(p.Title)[:previewTitleMax])) + "…"
	}
	if t.Estimate.Valid {
		p.Estimate = formatDuration(int(t.Estimate.Int64))
	}
	if t.Category.Valid {
		p.Category = t.Category.String
	}
	if t.DueDate.Valid {
		p.Due = formatDueDate(t.DueDate.String, today)
		p.Upcoming = view != "" && t.DueDate.String > view
	}
	return p
}
