package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
)

const (
	cellSize   = 11
	cellGap    = 2
	cellStep   = cellSize + cellGap
	leftMargin = 32
	topMargin  = 20
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

type cell struct {
	X, Y  int
	Level int
	Date  string
	Label string
	Value float64
	Best  bool
}

type textLabel struct {
	X, Y int
	Text string
}

type swatch struct {
	X, Y  int
	Level int
}

type heatmapData struct {
	Width    int
	Height   int
	CellSize int
	Cells    []cell
	Months   []textLabel
	Days     []textLabel
	Legend   []swatch
	LessX    int
	MoreX    int
	LegendY  int
	Detail   bool
}

func buildHeatmap(totals []store.DailyTotal, year int, bestDate string) heatmapData {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	start = start.AddDate(0, 0, -((int(start.Weekday()) + 6) % 7))
	end := time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	values := make(map[string]float64, len(totals))
	for _, t := range totals {
		values[t.LocalDate] = t.Value
	}
	lastMonth := time.Month(0)
	week := 0

	var cells []cell
	var months []textLabel
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		row := (int(d.Weekday()) + 6) % 7
		if row == 0 && d.After(start) {
			week++
		}
		if d.Year() != year {
			continue
		}
		if d.Month() != lastMonth && row < 3 {
			lastMonth = d.Month()
			months = append(months, textLabel{X: leftMargin + week*cellStep, Y: 14, Text: d.Format("Jan")})
		}
		date := d.Format("2006-01-02")
		v := values[date]
		cells = append(cells, cell{
			X:     leftMargin + week*cellStep,
			Y:     topMargin + row*cellStep,
			Level: levelFor(v),
			Date:  date,
			Label: d.Format("Mon, 2 Jan 2006"),
			Value: v,
			Best:  date == bestDate,
		})
	}
	width := leftMargin + (week+1)*cellStep
	height := topMargin + 7*cellStep + 28
	legendY := height - 6
	legendX := width - 130
	var legend []swatch
	for i := 0; i < levels; i++ {
		legend = append(legend, swatch{
			X:     legendX + 30 + i*cellStep,
			Y:     legendY - 9,
			Level: i,
		})
	}
	var days []textLabel
	for i, name := range []string{"Mon", "Wed", "Fri"} {
		days = append(days, textLabel{
			X:    0,
			Y:    topMargin + i*2*cellStep + 9,
			Text: name,
		})
	}

	return heatmapData{
		Width:    width,
		Height:   height,
		CellSize: cellSize,
		Cells:    cells,
		Months:   months,
		Days:     days,
		Legend:   legend,
		LessX:    legendX,
		MoreX:    legendX + 36 + 5*cellStep,
		LegendY:  legendY,
	}
}

func (s *Server) handleHeatmap(w http.ResponseWriter, r *http.Request) {
	eventType := r.URL.Query().Get("type")
	if eventType == "" {
		eventType = "contributions"
	}
	year := time.Now().Year()
	if y, err := strconv.Atoi(r.URL.Query().Get("year")); err == nil {
		year = y
	}
	view, err := s.view(r.Context(), eventType, year)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	s.renderPartial(w, "view.html", view)
}
