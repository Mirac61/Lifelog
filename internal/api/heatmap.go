package api

import (
	"time"

	"github.com/Mirac61/lifelog/internal/store"
	"github.com/Mirac61/lifelog/web/templ/git"
)

const (
	cellSize   = 11
	cellGap    = 2
	cellStep   = cellSize + cellGap
	leftMargin = 32
	topMargin  = 20

	// Sized to the best-day label's width so the SVG edge lands right after it.
	bestLegendWidth = 106
)

// CSS owns the ramp colours (.l0-.l4 in style.css).
const levels = 5

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

func buildHeatmap(totals []store.DailyTotal, year int, bestDate string) git.HeatmapData {
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

	var cells []git.Cell
	var months []git.TextLabel
	for d := start; !d.After(renderEnd); d = d.AddDate(0, 0, 1) {
		row := (int(d.Weekday()) + 6) % 7
		if row == 0 && d.After(start) {
			week++
		}
		if d.Year() != year {
			continue
		}
		// Anchor to the 1st's week so month-label spacing stays even.
		if d.Month() != lastMonth {
			lastMonth = d.Month()
			months = append(months, git.TextLabel{X: leftMargin + week*cellStep, Y: 14, Text: d.Format("Jan")})
		}
		date := d.Format("2006-01-02")
		v := values[date]
		cells = append(cells, git.Cell{
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
	// 138 fits "Wenig"; offsets below shift with it.
	legendX := width - 138

	// Second legend entry (not a ramp rung), only when there's a best day.
	hasBest := bestDate != ""
	if hasBest {
		width += bestLegendWidth
	}

	var legend []git.Swatch
	for i := range levels {
		legend = append(legend, git.Swatch{
			X:     legendX + 38 + i*cellStep,
			Y:     legendY - 9,
			Level: i,
		})
	}
	var days []git.TextLabel
	for i, name := range []string{"Mon", "Wed", "Fri", "Sun"} {
		days = append(days, git.TextLabel{
			X:    0,
			Y:    topMargin + i*2*cellStep + 9,
			Text: name,
		})
	}

	moreX := legendX + 44 + 5*cellStep
	dividerX := moreX + 34
	bestSwatchX := dividerX + 20
	bestLabelX := bestSwatchX + cellSize + 7

	return git.HeatmapData{
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
