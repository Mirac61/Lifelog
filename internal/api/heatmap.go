package api

import (
	"fmt"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
	"github.com/Mirac61/lifelog/web/templ/git"
	"github.com/Mirac61/lifelog/web/templ/ui"
)

// levelFor maps a day onto the design's four-step ramp. Level 0 is a day with
// nothing on it, which the grid draws as a hairline square.
func levelFor(v float64) int {
	switch {
	case v <= 0:
		return 0
	case v <= 3:
		return 1
	case v <= 8:
		return 2
	case v <= 15:
		return 3
	default:
		return 4
	}
}

// mondayIndex is the row a day sits in: the design's grid starts on Monday.
func mondayIndex(d time.Time) int { return (int(d.Weekday()) + 6) % 7 }

// buildHeatmap fills the year column by column, one column per week. Days
// outside the year — and days still ahead of today — stay void, so the current
// year ends where the data ends instead of trailing off into empty squares.
func buildHeatmap(totals []store.DailyTotal, year int, stats store.Stats, label string) git.Heatmap {
	values := make(map[string]float64, len(totals))
	for _, t := range totals {
		values[t.LocalDate] = t.Value
	}

	jan1 := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	dec31 := time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	start := jan1.AddDate(0, 0, -mondayIndex(jan1))
	end := dec31.AddDate(0, 0, 6-mondayIndex(dec31))
	cutoff := dec31
	if now := time.Now(); year == now.Year() {
		cutoff = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}

	hm := git.Heatmap{
		Label: fmt.Sprintf("%s %s an %d Tagen in %d", deNumber(int(stats.Total)), label, stats.ActiveDays, year),
	}
	month := time.Month(0)
	week := 1
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if mondayIndex(d) == 0 && d.After(start) {
			week++
		}
		if d.Year() != year || d.After(cutoff) {
			hm.Cells = append(hm.Cells, git.HeatCell{Void: true})
			continue
		}
		if d.Month() != month {
			if len(hm.Months) > 0 {
				hm.Months[len(hm.Months)-1].To = week
			}
			month = d.Month()
			hm.Months = append(hm.Months, git.MonthSpan{Label: ui.MonthShort[month], From: week})
		}
		date := d.Format("2006-01-02")
		value := values[date]
		cell := git.HeatCell{Level: levelFor(value), Best: date == stats.BestDate}
		if value > 0 {
			cell.Date = date
			cell.Tip = fmt.Sprintf("%s · %s %s", ui.DayMonthYear(d), deNumber(int(value)), label)
		}
		hm.Cells = append(hm.Cells, cell)
	}
	hm.Weeks = week
	if len(hm.Months) > 0 {
		hm.Months[len(hm.Months)-1].To = week + 1
	}
	return hm
}

// buildMonths sums the year into the twelve bars of the "Nach Monat" pane.
// Heights are px against the design's 72px peak.
func buildMonths(totals []store.DailyTotal, year int) []git.MonthBar {
	var sums [13]float64
	for _, t := range totals {
		d, err := time.Parse("2006-01-02", t.LocalDate)
		if err != nil || d.Year() != year {
			continue
		}
		sums[d.Month()] += t.Value
	}
	var peak float64
	for _, v := range sums {
		peak = max(peak, v)
	}

	months := make([]git.MonthBar, 0, 12)
	for m := time.January; m <= time.December; m++ {
		height := 2
		if peak > 0 {
			height = max(2, int(sums[m]/peak*72+0.5))
		}
		months = append(months, git.MonthBar{
			Label:  ui.MonthShort[m],
			Value:  deNumber(int(sums[m])),
			Height: height,
			Peak:   sums[m] > 0 && sums[m] == peak,
		})
	}
	return months
}
