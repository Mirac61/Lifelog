package api

import (
	"time"

	"github.com/Mirac61/lifelog/internal/store"
	views "github.com/Mirac61/lifelog/web/templ"
)

const (
	cellSize   = 11
	cellGap    = 2
	cellStep   = cellSize + cellGap
	leftMargin = 32
	topMargin  = 20

	// Room past the ramp for the divider and best-day swatch label, sized
	// to the label's own width so the SVG edge lands right after the text.
	bestLegendWidth = 106
)

// levels is the number of ramp steps; CSS owns the colours behind them
// (see the .l0-.l4 rules in style.css, mixed from --accent).
const levels = 5

// levelFor buckets a day's commit count into a ramp step. Thresholds live
// here because they are about the data; the colours do not, because they are
// about the domain the page is showing.
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

func buildHeatmap(totals []store.DailyTotal, year int, bestDate string) views.HeatmapData {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	start = start.AddDate(0, 0, -((int(start.Weekday()) + 6) % 7))
	end := time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	cutoff := end
	if now := time.Now(); year == now.Year() {
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if today.Before(cutoff) {
			cutoff = today
		}
	}
	renderEnd := end
	if buffered := cutoff.AddDate(0, 0, 7); buffered.Before(renderEnd) {
		buffered = buffered.AddDate(0, 0, 6-((int(buffered.Weekday())+6)%7))
		if buffered.Before(renderEnd) {
			renderEnd = buffered
		}
	}
	values := make(map[string]float64, len(totals))
	for _, t := range totals {
		values[t.LocalDate] = t.Value
	}
	lastMonth := time.Month(0)
	week := 0

	var cells []views.Cell
	var months []views.TextLabel
	for d := start; !d.After(renderEnd); d = d.AddDate(0, 0, 1) {
		row := (int(d.Weekday()) + 6) % 7
		if row == 0 && d.After(start) {
			week++
		}
		if d.Year() != year {
			continue
		}
		// Anchor to the week containing the 1st, regardless of weekday, so
		// spacing between month labels stays even.
		if d.Month() != lastMonth {
			lastMonth = d.Month()
			months = append(months, views.TextLabel{X: leftMargin + week*cellStep, Y: 14, Text: d.Format("Jan")})
		}
		date := d.Format("2006-01-02")
		v := values[date]
		cells = append(cells, views.Cell{
			X:      leftMargin + week*cellStep,
			Y:      topMargin + row*cellStep,
			Level:  levelFor(v),
			Date:   date,
			Label:  d.Format("Mon, 2 Jan 2006"),
			Value:  v,
			Best:   date == bestDate,
			Future: d.After(cutoff),
		})
	}
	width := leftMargin + (week+1)*cellStep
	height := topMargin + 7*cellStep + 28
	legendY := height - 6
	// Wider than a plain "Less"/"More" ramp needs, to fit "Wenig"; the
	// swatch and moreX offsets below shift by the same amount.
	legendX := width - 138

	// A second, categorical legend entry past a divider, not another rung
	// on the intensity ramp — only present when the year has a best day.
	hasBest := bestDate != ""
	if hasBest {
		width += bestLegendWidth
	}

	var legend []views.Swatch
	for i := range levels {
		legend = append(legend, views.Swatch{
			X:     legendX + 38 + i*cellStep,
			Y:     legendY - 9,
			Level: i,
		})
	}
	var days []views.TextLabel
	for i, name := range []string{"Mon", "Wed", "Fri", "Sun"} {
		days = append(days, views.TextLabel{
			X:    0,
			Y:    topMargin + i*2*cellStep + 9,
			Text: name,
		})
	}

	moreX := legendX + 44 + 5*cellStep
	dividerX := moreX + 34
	bestSwatchX := dividerX + 20
	bestLabelX := bestSwatchX + cellSize + 7

	return views.HeatmapData{
		Width:       width,
		Height:      height,
		CellSize:    cellSize,
		Cells:       cells,
		Months:      months,
		Days:        days,
		Legend:      legend,
		LessX:       legendX,
		MoreX:       moreX,
		LegendY:     legendY,
		HasBest:     hasBest,
		DividerX:    dividerX,
		BestSwatchX: bestSwatchX,
		BestLabelX:  bestLabelX,
	}
}
